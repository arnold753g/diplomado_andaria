package config

import "testing"

func TestProxyRequiresExplicitNonWildcardNetworks(t *testing.T) {
	validEnvironment(t)
	t.Setenv("TRUST_PROXY", "true")
	for _, value := range []string{"", "0.0.0.0/0", "::/0", "invalid"} {
		t.Setenv("TRUSTED_PROXY_CIDRS", value)
		if _, err := Load(); err == nil {
			t.Fatalf("unsafe proxy network accepted: %q", value)
		}
	}
	t.Setenv("TRUSTED_PROXY_CIDRS", "192.0.2.10/32,2001:db8::1/128")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.TrustProxy || len(cfg.TrustedProxyCIDRs) != 2 {
		t.Fatal("trusted peers lost")
	}
}

func TestProductionRequiresVerificationButCanWaitForMailCredentials(t *testing.T) {
	validEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_COOKIE_SECURE", "true")
	t.Setenv("ALLOWED_ORIGINS", "https://andaria.example")
	t.Setenv("FRONTEND_URL", "https://andaria.example")
	t.Setenv("REGISTRATION_ENABLED", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EmailVerificationRequired || cfg.SMTPHost != "" {
		t.Fatal("production verification default missing")
	}
	t.Setenv("EMAIL_VERIFICATION_REQUIRED", "false")
	if _, err := Load(); err == nil {
		t.Fatal("unverified public registration enabled in production")
	}
	t.Setenv("EMAIL_VERIFICATION_REQUIRED", "true")
	t.Setenv("SMTP_HOST", "smtp.example.test")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_FROM", "Sender <sender@example.test>")
	if _, err := Load(); err == nil {
		t.Fatal("unsupported mail sender accepted")
	}
	t.Setenv("SMTP_FROM", "sender@example.test")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DB_AUTO_MIGRATE", "true")
	if _, err := Load(); err == nil {
		t.Fatal("API migrations enabled in production")
	}
}

func TestDatabaseCredentialsAreQuotedAsSingleDSNValues(t *testing.T) {
	if got := dsnValue("a b'c\\d"); got != "'a b\\'c\\\\d'" {
		t.Fatal("DSN quoting changed")
	}
}
