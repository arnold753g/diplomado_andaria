package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"starter-backend/internal/models"
)

type contextKey string

const principalKey contextKey = "principal"

type Principal struct {
	User      models.User
	SessionID string
}

func WithPrincipal(r *http.Request, principal Principal) *http.Request {
	ctx := context.WithValue(r.Context(), principalKey, principal)
	return r.WithContext(ctx)
}

func CurrentPrincipal(r *http.Request) (Principal, bool) {
	principal, ok := r.Context().Value(principalKey).(Principal)
	return principal, ok
}

func ClientIP(r *http.Request, trustProxy bool) string {
	trusted, _ := r.Context().Value(trustedProxyKey).(bool)
	if trustProxy && trusted {
		candidates := []string{r.Header.Get("X-Real-IP")}
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			first, _, _ := strings.Cut(forwarded, ",")
			candidates = append(candidates, first)
		}
		for _, candidate := range candidates {
			candidate = strings.TrimSpace(candidate)
			if ip := net.ParseIP(candidate); ip != nil {
				return ip.String()
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
