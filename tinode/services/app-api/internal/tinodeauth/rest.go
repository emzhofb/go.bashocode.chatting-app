package tinodeauth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ikhda/tinode-chat/services/app-api/internal/auth"
)

type Handler struct {
	Service auth.Service
}

type request struct {
	Endpoint string  `json:"endpoint"`
	Secret   string  `json:"secret"`
	Addr     string  `json:"addr"`
	Rec      *record `json:"rec"`
}

type record struct {
	UID     string `json:"uid,omitempty"`
	AuthLvl string `json:"authlvl,omitempty"`
	State   string `json:"state,omitempty"`
}

type response struct {
	Err    string   `json:"err,omitempty"`
	Rec    *record  `json:"rec,omitempty"`
	Bool   *bool    `json:"boolval,omitempty"`
	StrArr []string `json:"strarr,omitempty"`
	NewAcc *newAcc  `json:"newacc,omitempty"`
}

type newAcc struct {
	Auth   string         `json:"auth"`
	Anon   string         `json:"anon"`
	Public map[string]any `json:"public,omitempty"`
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeResponse(w, response{Err: "malformed"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32*1024)
	var input request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		writeResponse(w, response{Err: "malformed"})
		return
	}
	if input.Endpoint == "" {
		input.Endpoint = endpointFromPath(r.URL.Path)
	}
	switch input.Endpoint {
	case "auth":
		h.auth(w, r, input)
	case "link":
		h.link(w, r, input)
	case "checkunique":
		h.checkUnique(w, r, input)
	case "rtagns":
		writeResponse(w, response{StrArr: []string{"basic", "email"}})
	case "add", "del", "gen", "upd":
		writeResponse(w, response{Err: "unsupported"})
	default:
		writeResponse(w, response{Err: "unsupported"})
	}
}

func (h Handler) auth(w http.ResponseWriter, r *http.Request, input request) {
	username, password, err := decodeSecret(input.Secret)
	if err != nil {
		writeResponse(w, response{Err: "malformed"})
		return
	}
	user, uid, err := h.Service.AuthenticateTinode(r.Context(), username, password)
	if err != nil {
		writeResponse(w, response{Err: "failed"})
		return
	}
	rec := &record{AuthLvl: "auth", State: "ok"}
	if uid != "" {
		rec.UID = uid
		writeResponse(w, response{Rec: rec})
		return
	}
	writeResponse(w, response{Rec: rec, NewAcc: &newAcc{Auth: "JRWPS", Anon: "N", Public: map[string]any{"display_name": user.DisplayName, "avatar_url": user.AvatarURL}}})
}

func (h Handler) link(w http.ResponseWriter, r *http.Request, input request) {
	username, password, err := decodeSecret(input.Secret)
	if err != nil || input.Rec == nil || input.Rec.UID == "" {
		writeResponse(w, response{Err: "malformed"})
		return
	}
	if _, _, err := h.Service.AuthenticateTinode(r.Context(), username, password); err != nil {
		writeResponse(w, response{Err: "failed"})
		return
	}
	if err := h.Service.LinkTinode(r.Context(), username, input.Rec.UID); err != nil {
		if strings.Contains(err.Error(), "conflict") {
			writeResponse(w, response{Err: "duplicate value"})
			return
		}
		writeResponse(w, response{Err: "internal"})
		return
	}
	writeResponse(w, response{Rec: &record{UID: input.Rec.UID, AuthLvl: "auth", State: "ok"}})
}

func (h Handler) checkUnique(w http.ResponseWriter, r *http.Request, input request) {
	username, _, err := decodeSecret(input.Secret)
	if err != nil {
		writeResponse(w, response{Err: "malformed"})
		return
	}
	normalized, err := auth.NormalizeUsername(username)
	if err != nil {
		writeResponse(w, response{Err: "policy"})
		return
	}
	available, err := h.Service.TinodeUsernameAvailable(r.Context(), normalized)
	if err != nil {
		writeResponse(w, response{Err: "internal"})
		return
	}
	writeResponse(w, response{Bool: &available})
}

func decodeSecret(value string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.Contains(parts[0], "\x00") {
		return "", "", errors.New("invalid secret")
	}
	return parts[0], parts[1], nil
}

func endpointFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func writeResponse(w http.ResponseWriter, value response) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
