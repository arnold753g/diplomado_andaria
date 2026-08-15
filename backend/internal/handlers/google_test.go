package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"starter-backend/internal/config"
	"strings"
	"testing"
)

type googleTransport func(*http.Request) (*http.Response, error)

func (f googleTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGoogleExchangeRequiresVerifiedEmail(t *testing.T) {
	for _, verified := range []bool{true, false} {
		h, err := NewAuthHandler(nil, config.Config{PasswordBcryptCost: 10, GoogleClientID: "test-client", GoogleClientSecret: "test-secret", GoogleRedirectURL: "http://localhost/callback"})
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		h.googleClient = &http.Client{Transport: googleTransport(func(r *http.Request) (*http.Response, error) {
			calls++
			body := `{"access_token":"access","token_type":"Bearer"}`
			if calls == 1 {
				if r.URL.String() != "https://oauth2.googleapis.com/token" || r.Method != "POST" {
					t.Fatal("unexpected token endpoint")
				}
				if err := r.ParseForm(); err != nil {
					t.Fatal(err)
				}
				if r.Form.Get("code_verifier") != "verifier" || r.Form.Get("client_id") != "test-client" {
					t.Fatal("missing code binding")
				}
			} else {
				if r.URL.String() != "https://openidconnect.googleapis.com/v1/userinfo" || r.Header.Get("Authorization") != "Bearer access" {
					t.Fatal("userinfo must be authenticated")
				}
				body = `{"sub":"subject","email":"tourist@example.test","email_verified":false}`
				if verified {
					body = strings.Replace(body, "false", "true", 1)
				}
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})}
		identity, err := h.exchangeGoogle(context.Background(), "code", "verifier")
		if verified && (err != nil || identity.Subject != "subject") {
			t.Fatalf("verified identity rejected: %v", err)
		}
		if !verified && err == nil {
			t.Fatal("unverified email accepted")
		}
	}
}
func TestGoogleRejectsInvalidStateBeforeDatabaseOrProvider(t *testing.T) {
	h, err := NewAuthHandler(nil, config.Config{PasswordBcryptCost: 10, GoogleClientID: "test", FrontendURL: "http://localhost:3000"})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "/api/v1/auth/google/callback?state=forged&code=forged", nil)
	w := httptest.NewRecorder()
	h.GoogleCallback(w, r)
	if w.Code != 303 || !strings.Contains(w.Header().Get("Location"), "invalid_state") {
		t.Fatalf("invalid state was not rejected: %d", w.Code)
	}
}
