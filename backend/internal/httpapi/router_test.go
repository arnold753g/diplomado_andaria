package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"starter-backend/internal/config"
)

func testConfig() config.Config {
	return config.Config{AllowedOrigins: []string{"http://localhost:3000"}, SessionSecret: "0123456789abcdef0123456789abcdef",
		SessionTTL: time.Hour, SessionIdleTimeout: 15 * time.Minute, PasswordBcryptCost: 10,
		AuthRatePerMinute: 10, APIRatePerMinute: 100}
}

func TestHealth(t *testing.T) {
	handler, err := New(nil, testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestProtectedRouteRejectsAnonymousUser(t *testing.T) {
	handler, err := New(nil, testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
