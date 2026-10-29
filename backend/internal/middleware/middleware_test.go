package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"starter-backend/internal/models"
)

func TestRecoveryDoesNotExposeInternalDetails(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("private SQL connection details")
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/failure", nil))
	body := response.Body.String()
	if response.Code != 500 || !strings.Contains(body, "INTERNAL_ERROR") || strings.Contains(body, "private SQL") || strings.Contains(body, "goroutine") {
		t.Fatalf("unsafe internal error: status=%d body=%s", response.Code, body)
	}
}

func TestRoleMiddlewareRejectsCrossRoleAccess(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := RequireRoles(models.RoleAdmin)(next)
	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request = WithPrincipal(request, Principal{User: models.User{Role: models.RoleUser}})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", response.Code, http.StatusForbidden)
	}
}

func TestClientIPIgnoresForwardedHeaderWhenProxyIsUntrusted(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "127.0.0.1:1234"
	request.Header.Set("X-Forwarded-For", "203.0.113.9")
	if got := ClientIP(request, false); got != "127.0.0.1" {
		t.Fatalf("client IP=%q", got)
	}
}
