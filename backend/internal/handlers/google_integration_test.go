package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"starter-backend/internal/config"
	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
	"strings"
	"testing"
	"time"
)

func TestGoogleIdentityCreationAndExplicitLink(t *testing.T) {
	db := testutil.Database(t)
	h, err := NewAuthHandler(db, config.Config{PasswordBcryptCost: 10, RegistrationEnabled: true, SessionIdleTimeout: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	identity := googleIdentity{Subject: "google-1", Email: "google@example.test", EmailVerified: true, GivenName: "Ana", FamilyName: "Prueba"}
	user, err := h.resolveGoogleUser(identity, oauthAttempt{})
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != models.RoleUser || user.PasswordHash != "" || user.GoogleSubject == nil {
		t.Fatal("Google must create a tourist without a local password")
	}
	existing, err := h.resolveGoogleUser(identity, oauthAttempt{})
	if err != nil || existing.ID != user.ID {
		t.Fatal("Google login duplicated the user")
	}
	identity.Subject = "another-subject"
	if _, err := h.resolveGoogleUser(identity, oauthAttempt{}); err == nil || err.Error() != "email_exists" {
		t.Fatal("same email auto-linked")
	}
	hash, _ := security.HashPassword("test-password-123", 10)
	local := models.User{Email: "local@example.test", PasswordHash: hash, FirstName: "Local", LastName: "User", Role: models.RoleAgency, Status: models.StatusActive}
	if err := db.Create(&local).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	session := models.Session{ID: "link-session", UserID: local.ID, TokenHash: "link-token", LastActivityAt: now, ExpiresAt: now.Add(time.Hour)}
	if err := db.Create(&session).Error; err != nil {
		t.Fatal(err)
	}
	identity.Subject = "google-local"
	identity.Email = local.Email
	attempt := oauthAttempt{UserID: &local.ID, SessionID: &session.ID}
	linked, err := h.resolveGoogleUser(identity, attempt)
	if err != nil || linked.ID != local.ID {
		t.Fatalf("explicit linking failed: %v", err)
	}
	db.First(&local, local.ID)
	if local.Role != models.RoleAgency || local.GoogleSubject == nil {
		t.Fatal("linking changed role or was not persisted")
	}
	db.Model(&local).Update("status", models.StatusInactive)
	if _, err := h.resolveGoogleUser(identity, oauthAttempt{}); err == nil || err.Error() != "inactive" {
		t.Fatal("inactive Google account logged in")
	}
	h.cfg.RegistrationEnabled = false
	identity.Subject = "new-google"
	identity.Email = "new@example.test"
	if _, err := h.resolveGoogleUser(identity, oauthAttempt{}); err == nil || err.Error() != "registration_disabled" {
		t.Fatal("Google bypassed registration setting")
	}
}

func TestGoogleCallbackConsumesStateOnce(t *testing.T) {
	db := testutil.Database(t)
	h, err := NewAuthHandler(db, config.Config{PasswordBcryptCost: 10, GoogleClientID: "client", GoogleRedirectURL: "http://localhost/callback", FrontendURL: "http://localhost:3000", RegistrationEnabled: true, SessionTTL: time.Hour, SessionIdleTimeout: time.Hour, SessionSecret: strings.Repeat("s", 32)})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	h.googleClient = &http.Client{Transport: googleTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		body := `{"access_token":"access","token_type":"Bearer"}`
		if strings.Contains(r.URL.Path, "userinfo") {
			body = `{"sub":"callback-google","email":"callback@example.test","email_verified":true,"given_name":"Ana","family_name":"Prueba"}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	start := httptest.NewRecorder()
	h.GoogleStart(start, httptest.NewRequest("POST", "/auth/google/start", nil))
	if start.Code != 200 {
		t.Fatal(start.Body.String())
	}
	cookie := start.Result().Cookies()[0]
	callback := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/auth/google/callback?state="+cookie.Value+"&code=code", nil)
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		h.GoogleCallback(w, r)
		return w
	}
	first := callback()
	if !strings.Contains(first.Header().Get("Location"), "success") || calls != 2 {
		t.Fatalf("callback failed: %s", first.Header().Get("Location"))
	}
	second := callback()
	if !strings.Contains(second.Header().Get("Location"), "invalid_state") || calls != 2 {
		t.Fatal("OAuth callback replay was accepted")
	}
}
