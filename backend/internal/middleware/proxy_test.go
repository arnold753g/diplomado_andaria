package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyTrustRequiresVerifiedImmediatePeer(t *testing.T) {
	for _, tc := range []struct {
		peer    string
		enabled bool
		want    string
	}{
		{"192.0.2.10:8080", true, "203.0.113.5"},
		{"192.0.2.11:8080", true, "192.0.2.11"},
		{"192.0.2.10:8080", false, "192.0.2.10"},
	} {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = tc.peer
		r.Header.Set("X-Real-IP", "203.0.113.5")
		r.Header.Set("X-Forwarded-For", "198.51.100.9")
		handler := TrustedProxies(tc.enabled, []string{"192.0.2.10/32"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := ClientIP(r, true); got != tc.want {
				t.Errorf("peer=%s enabled=%v got=%s want=%s", tc.peer, tc.enabled, got, tc.want)
			}
		}))
		handler.ServeHTTP(httptest.NewRecorder(), r)
	}
}
