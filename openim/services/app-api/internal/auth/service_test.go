package auth

import "testing"

func TestOpenIMUserID(t *testing.T) {
	if got, want := openIMUserID("27a1b92d-c31c-4472-a6ee-fe362412c0bf"), "app_27a1b92dc31c4472a6eefe362412c0bf"; got != want {
		t.Fatalf("openIMUserID() = %q, want %q", got, want)
	}
}
