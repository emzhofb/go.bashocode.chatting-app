package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/ikhda/openim-chat/services/app-api/internal/auth"
	"github.com/ikhda/openim-chat/services/app-api/internal/openim"
)

type contextKey string

const userContextKey contextKey = "authenticated-user"

type AuthHandler struct {
	Service      auth.Service
	OpenIM       *openim.Client
	PublicAPIURL string
	PublicWSURL  string
	PlatformID   int
	Limiter      *RateLimiter
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
	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{"user": user, "session": session}})
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
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Username/email atau password salah", nil)
		} else {
			writeAuthError(w, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"user": user, "session": session}})
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
	if user, ok := r.Context().Value(userContextKey).(auth.User); ok && h.OpenIM != nil {
		if err := h.OpenIM.ForceLogout(r.Context(), user.OpenIMUserID, h.PlatformID); err != nil {
			writeError(w, http.StatusBadGateway, "OPENIM_ERROR", "Sesi OpenIM belum dapat dicabut", nil)
			return
		}
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

func (h AuthHandler) OpenIMSession(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(auth.User)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Login diperlukan", nil)
		return
	}
	if h.OpenIM == nil {
		writeError(w, http.StatusServiceUnavailable, "OPENIM_UNAVAILABLE", "OpenIM belum tersedia", nil)
		return
	}
	token, err := h.OpenIM.GetUserToken(r.Context(), user.OpenIMUserID, h.PlatformID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "OPENIM_ERROR", "Gagal membuat sesi OpenIM", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"user_id": user.OpenIMUserID, "token": token.Token, "expire_time_seconds": token.ExpireTimeSeconds, "api_addr": h.PublicAPIURL, "ws_addr": h.PublicWSURL}})
}

func (h AuthHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(auth.User)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Login diperlukan", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len([]rune(query)) < 2 || len([]rune(query)) > 64 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Query pencarian harus 2-64 karakter", map[string]string{"q": "invalid length"})
		return
	}
	limit := 20
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > 50 {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Limit harus 1-50", map[string]string{"limit": "invalid value"})
			return
		}
		limit = parsed
	}
	users, err := h.Service.SearchUsers(r.Context(), query, user.ID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Pencarian user gagal", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"users": users}})
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
	key := endpoint + ":unknown"
	if ip := remoteIP(r); ip != nil {
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
	if strings.Contains(err.Error(), "provision OpenIM") || strings.Contains(err.Error(), "OpenIM") {
		writeError(w, http.StatusBadGateway, "OPENIM_ERROR", "Gagal menyiapkan akun messaging", nil)
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
