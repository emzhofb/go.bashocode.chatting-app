package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHandlerAllowsConfiguredOrigin(t *testing.T) {
	handler := CORSHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }), "http://localhost:3000")
	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent || res.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("unexpected CORS response: status=%d headers=%v", res.Code, res.Header())
	}
}

func TestCORSHandlerRejectsUnknownOrigin(t *testing.T) {
	handler := CORSHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }), "http://localhost:3000")
	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
	req.Header.Set("Origin", "http://evil.test")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want %d", res.Code, http.StatusForbidden)
	}
}
