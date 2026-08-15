package config

import (
	"testing"
	"time"
)

func validEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DB_PASSWORD", "database-test-password")
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
	t.Setenv("APP_ENV", "test")
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_REDIRECT_URL", "")
	t.Setenv("FRONTEND_URL", "http://localhost:3000")
}

func TestGoogleConfigurationRequiresCompleteTrustedURLs(t *testing.T) {
	validEnvironment(t)
	t.Setenv("GOOGLE_CLIENT_ID", "client")
	if _, err := Load(); err == nil {
		t.Fatal("partial Google config accepted")
	}
	t.Setenv("GOOGLE_CLIENT_SECRET", "secret")
	t.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FRONTEND_URL", "https://untrusted.example")
	if _, err := Load(); err == nil {
		t.Fatal("untrusted frontend accepted")
	}
	t.Setenv("FRONTEND_URL", "http://localhost:3000")
	t.Setenv("GOOGLE_REDIRECT_URL", "http://public.example/callback")
	if _, err := Load(); err == nil {
		t.Fatal("insecure remote callback accepted")
	}
}

func TestLoadRejectsMissingSecret(t *testing.T) {
	validEnvironment(t)
	t.Setenv("SESSION_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing SESSION_SECRET to fail")
	}
}

func TestLoadParsesSessionDurations(t *testing.T) {
	validEnvironment(t)
	t.Setenv("SESSION_TTL", "8h")
	t.Setenv("SESSION_IDLE_TIMEOUT", "20m")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SessionTTL != 8*time.Hour || cfg.SessionIdleTimeout != 20*time.Minute {
		t.Fatalf("unexpected durations: %v %v", cfg.SessionTTL, cfg.SessionIdleTimeout)
	}
}

func TestProductionRequiresSecureCookie(t *testing.T) {
	validEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	if _, err := Load(); err == nil {
		t.Fatal("expected insecure production cookie to fail")
	}
}
