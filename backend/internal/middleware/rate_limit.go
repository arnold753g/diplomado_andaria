package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"starter-backend/internal/respond"
)

type rateEntry struct {
	count     int
	windowEnd time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	entries    map[string]*rateEntry
	limit      int
	maxEntries int
	trustProxy bool
	now        func() time.Time
}

func NewRateLimiter(limit int, trustProxy bool) *RateLimiter {
	return &RateLimiter{entries: make(map[string]*rateEntry), limit: limit, maxEntries: 10000, trustProxy: trustProxy, now: time.Now}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := l.now()
		key := ClientIP(r, l.trustProxy)
		l.mu.Lock()
		entry, ok := l.entries[key]
		if !ok && len(l.entries) >= l.maxEntries {
			for candidate, value := range l.entries {
				if !now.Before(value.windowEnd) {
					delete(l.entries, candidate)
				}
			}
			if len(l.entries) >= l.maxEntries {
				l.mu.Unlock()
				w.Header().Set("Retry-After", "60")
				respond.Error(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests; try again later", nil)
				return
			}
		}
		if !ok || !now.Before(entry.windowEnd) {
			entry = &rateEntry{windowEnd: now.Add(time.Minute)}
			l.entries[key] = entry
		}
		entry.count++
		count := entry.count
		reset := entry.windowEnd
		l.mu.Unlock()

		remaining := l.limit - count
		if remaining < 0 {
			remaining = 0
		}
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if count > l.limit {
			retry := int(reset.Sub(now).Seconds())
			if retry < 1 {
				retry = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			respond.Error(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests; try again later", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
