package httpapi

import (
	"net/http"
	"strings"
)

func CORSHandler(next http.Handler, allowed string) http.Handler {
	allowList := map[string]struct{}{}
	for _, origin := range strings.Split(allowed, ",") {
		if value := strings.TrimSpace(origin); value != "" {
			allowList[value] = struct{}{}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowList[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
