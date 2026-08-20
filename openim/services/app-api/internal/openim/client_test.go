package openim

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestProvisionAndIssueUserToken(t *testing.T) {
	adminCalls, registerCalls := 0, 0
	client := &Client{BaseURL: "http://openim.test", AdminUser: "imAdmin", AdminSecret: "server-secret"}
	client.HTTP = &http.Client{Transport: roundTripFunc(func(r *http.Request) *http.Response {
		if r.Header.Get("operationID") == "" {
			t.Error("operationID header is missing")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		response := func(status int, body string) *http.Response {
			return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}
		}
		switch r.URL.Path {
		case "/auth/get_admin_token":
			adminCalls++
			if body["secret"] != "server-secret" {
				t.Errorf("unexpected secret: %v", body["secret"])
			}
			return response(http.StatusOK, `{"errCode":0,"data":{"token":"admin-token","expireTimeSeconds":3600}}`)
		case "/user/get_users_info":
			if r.Header.Get("token") != "admin-token" {
				t.Errorf("missing admin token")
			}
			return response(http.StatusOK, `{"errCode":1004,"errMsg":"RecordNotFoundError"}`)
		case "/user/user_register":
			registerCalls++
			if r.Header.Get("token") != "admin-token" {
				t.Errorf("missing admin token")
			}
			return response(http.StatusOK, `{"errCode":0}`)
		case "/auth/get_user_token":
			if r.Header.Get("token") != "admin-token" {
				t.Errorf("missing admin token")
			}
			return response(http.StatusOK, `{"errCode":0,"data":{"token":"user-token","expireTimeSeconds":900}}`)
		default:
			return response(http.StatusNotFound, "")
		}
	})}
	if err := client.ProvisionUser(context.Background(), "user-1", "Alice"); err != nil {
		t.Fatalf("ProvisionUser() error = %v", err)
	}
	got, err := client.GetUserToken(context.Background(), "user-1", 5)
	if err != nil {
		t.Fatalf("GetUserToken() error = %v", err)
	}
	if got.Token != "user-token" || got.ExpireTimeSeconds != 900 {
		t.Fatalf("unexpected token: %+v", got)
	}
	if adminCalls != 1 || registerCalls != 1 {
		t.Fatalf("adminCalls=%d registerCalls=%d", adminCalls, registerCalls)
	}
}

func TestAPIErrorDoesNotExposeResponseBody(t *testing.T) {
	client := &Client{BaseURL: "http://openim.test", AdminUser: "imAdmin", AdminSecret: "server-secret", HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) *http.Response {
		return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("token=secret-value"))}
	})}}
	err := client.ProvisionUser(context.Background(), "user-1", "Alice")
	if err == nil || strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type roundTripFunc func(*http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req), nil }
