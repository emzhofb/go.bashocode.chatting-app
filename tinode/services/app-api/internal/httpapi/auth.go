package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ikhda/tinode-chat/services/app-api/internal/auth"
)

type contextKey string

const userContextKey contextKey = "authenticated-user"

type AuthHandler struct {
	Service auth.Service
	Limiter *RateLimiter
}

type registerRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.allow(w, r, "register") {
		return
	}
	var input registerRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", nil)
		return
	}
	user, session, err := h.Service.Register(r.Context(), auth.RegisterInput{Username: input.Username, Email: input.Email, Password: input.Password, DisplayName: input.DisplayName}, r.UserAgent(), remoteIP(r))
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{"user": user, "session": session, "tinode_login": map[string]string{"scheme": "basic", "username": user.Username}}})
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.allow(w, r, "login") {
		return
	}
	var input loginRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", nil)
		return
	}
	user, session, err := h.Service.Login(r.Context(), auth.LoginInput{Login: input.Login, Password: input.Password}, r.UserAgent(), remoteIP(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Username/email atau password salah", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"user": user, "session": session, "tinode_login": map[string]string{"scheme": "basic", "username": user.Username}}})
}

func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if !h.allow(w, r, "refresh") {
		return
	}
	var input refreshRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", nil)
		return
	}
	user, session, err := h.Service.Refresh(r.Context(), input.RefreshToken, r.UserAgent(), remoteIP(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Refresh token tidak valid", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"user": user, "session": session}})
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.Logout(r.Context(), bearerToken(r)); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal logout", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"logged_out": true}})
}

func (h AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(auth.User)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Login diperlukan", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"user": user}})
}

func (h AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.Service.UserByAccessToken(r.Context(), bearerToken(r))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Login diperlukan", nil)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	})
}

func bearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(value) < 7 || !strings.EqualFold(value[:6], "Bearer") || value[6] != ' ' {
		return ""
	}
	return strings.TrimSpace(value[7:])
}

func remoteIP(r *http.Request) net.IP {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return errors.New("request must contain one JSON object")
	} else if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func (h AuthHandler) allow(w http.ResponseWriter, r *http.Request, endpoint string) bool {
	ip := remoteIP(r)
	key := endpoint + ":unknown"
	if ip != nil {
		key = endpoint + ":" + ip.String()
	}
	if h.Limiter == nil || h.Limiter.Allow(key) {
		return true
	}
	w.Header().Set("Retry-After", "60")
	writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Terlalu banyak percobaan", nil)
	return false
}

func writeAuthError(w http.ResponseWriter, err error) {
	if field := auth.FieldName(err); field != "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", map[string]string{field: err.Error()})
		return
	}
	if strings.Contains(err.Error(), "already registered") {
		writeError(w, http.StatusConflict, "CONFLICT", "Username atau email sudah digunakan", nil)
		return
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan internal", nil)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": code, "message": message, "fields": fields}})
}
