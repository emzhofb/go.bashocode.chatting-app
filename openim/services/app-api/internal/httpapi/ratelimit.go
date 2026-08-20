package httpapi

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]bucket
}
type bucket struct {
	started time.Time
	count   int
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, buckets: make(map[string]bucket)}
}

func (l *RateLimiter) Allow(key string) bool {
	if l == nil || l.limit <= 0 {
		return true
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	current := l.buckets[key]
	if current.started.IsZero() || now.Sub(current.started) >= l.window {
		current = bucket{started: now}
	}
	if current.count >= l.limit {
		l.buckets[key] = current
		return false
	}
	current.count++
	l.buckets[key] = current
	return true
}
