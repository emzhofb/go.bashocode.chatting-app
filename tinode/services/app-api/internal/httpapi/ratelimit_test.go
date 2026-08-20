package httpapi

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	if !limiter.Allow("ip") || !limiter.Allow("ip") {
		t.Fatal("first two requests should be allowed")
	}
	if limiter.Allow("ip") {
		t.Fatal("third request should be rejected")
	}
	if !limiter.Allow("other-ip") {
		t.Fatal("different key should have its own bucket")
	}
}
