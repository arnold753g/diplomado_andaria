package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"starter-backend/internal/config"
	"starter-backend/internal/testutil"
)

func testConfig() config.Config {
	return config.Config{AllowedOrigins: []string{"http://localhost:3000"}, SessionSecret: "0123456789abcdef0123456789abcdef",
		SessionTTL: time.Hour, SessionIdleTimeout: 15 * time.Minute, PasswordBcryptCost: 10,
		AuthRatePerMinute: 10, APIRatePerMinute: 100}
}

func TestHealthRejectsMissingDatabase(t *testing.T) {
	handler, err := New(nil, testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/health", "/api/v1/health", "/api/v1/salud"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "SERVICE_UNAVAILABLE") {
			t.Fatalf("%s status=%d body=%s", path, response.Code, response.Body.String())
		}
	}
}

func TestHealthDatabaseReadiness(t *testing.T) {
	db := testutil.Database(t)
	handler, err := New(db, testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	check := func(want int) {
		t.Helper()
		for _, path := range []string{"/health", "/api/v1/health", "/api/v1/salud"} {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != want {
				t.Fatalf("%s status=%d body=%s", path, response.Code, response.Body.String())
			}
			if want == http.StatusOK && !strings.Contains(response.Body.String(), `"database":"ok"`) {
				t.Fatal("missing database readiness")
			}
			if strings.Contains(response.Body.String(), "SELECT") || strings.Contains(response.Body.String(), "sql:") {
				t.Fatal("internal details exposed")
			}
		}
	}
	check(http.StatusOK)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	check(http.StatusServiceUnavailable)
}

func TestHealthCancelledRequest(t *testing.T) {
	db := testutil.Database(t)
	handler, err := New(db, testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/salud", nil)
	ctx, cancel := context.WithCancel(r.Context())
	cancel()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, r.WithContext(ctx))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("cancelled request status=%d body=%s", response.Code, response.Body.String())
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
