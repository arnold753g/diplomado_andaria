package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterBoundsMemoryWithoutEvictingActiveLimits(t *testing.T) {
	limiter := NewRateLimiter(1, false)
	limiter.maxEntries = 2
	now := time.Now()
	limiter.now = func() time.Time { return now }
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	request := func(peer string, want int) {
		t.Helper()
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = peer + ":1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s got=%d want=%d", peer, w.Code, want)
		}
	}
	request("192.0.2.1", 204)
	request("192.0.2.2", 204)
	request("192.0.2.3", 429)
	request("192.0.2.1", 429)
	if len(limiter.entries) != 2 {
		t.Fatal("limiter memory grew beyond bound")
	}
	now = now.Add(time.Minute)
	request("192.0.2.3", 204)
}
