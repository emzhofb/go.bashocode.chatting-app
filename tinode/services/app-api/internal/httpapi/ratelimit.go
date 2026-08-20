package httpapi

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]rateEntry
	cleaned time.Time
}

type rateEntry struct {
	started time.Time
	count   int
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{limit: limit, window: window, entries: make(map[string]rateEntry)}
}

func (l *RateLimiter) Allow(key string) bool {
	if l == nil {
		return true
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cleaned.IsZero() || now.Sub(l.cleaned) >= l.window {
		for key, entry := range l.entries {
			if now.Sub(entry.started) >= l.window {
				delete(l.entries, key)
			}
		}
		l.cleaned = now
	}
	entry, ok := l.entries[key]
	if !ok || now.Sub(entry.started) >= l.window {
		l.entries[key] = rateEntry{started: now, count: 1}
		return true
	}
	if entry.count >= l.limit {
		return false
	}
	entry.count++
	l.entries[key] = entry
	return true
}
