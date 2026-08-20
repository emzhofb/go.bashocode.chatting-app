package tinodeauth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeSecret(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("alice:password"))
	username, password, err := decodeSecret(encoded)
	if err != nil || username != "alice" || password != "password" {
		t.Fatalf("got %q/%q/%v", username, password, err)
	}
}

func TestEndpointFromPath(t *testing.T) {
	if got := endpointFromPath("/internal/tinode/auth/checkunique"); got != "checkunique" {
		t.Fatalf("got %q", got)
	}
}

func TestUnsupportedEndpointResponse(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/internal/tinode/auth/add", strings.NewReader(`{"endpoint":"add"}`))
	recorder := httptest.NewRecorder()
	(Handler{}).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"err":"unsupported"`) {
		t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
	}
}
