package handlers

import (
	"bytes"
	"cloud.google.com/go/auth/credentials/idtoken"
	"context"
	"encoding/json"
	"net/http/httptest"
	"starter-backend/internal/config"
	"starter-backend/internal/models"
	"starter-backend/internal/testutil"
	"strings"
	"testing"
	"time"
)

func TestGoogleMobileClaimsRejectWrongAudienceIssuerExpiredAndUnverified(t *testing.T) {
	valid := func() *idtoken.Payload {
		return &idtoken.Payload{Audience: "andaria-web-client", Issuer: "https://accounts.google.com", Subject: "google-user-id", Expires: time.Now().Add(time.Hour).Unix(), Claims: map[string]any{"email": "turista@example.test", "email_verified": true, "given_name": "Ana", "family_name": "Paz"}}
	}
	if identity, err := googlePayloadIdentity(valid(), "andaria-web-client"); err != nil || identity.Subject != "google-user-id" {
		t.Fatalf("valid identity failed: %v", err)
	}
	cases := map[string]func(*idtoken.Payload){
		"audience":   func(p *idtoken.Payload) { p.Audience = "other-client" },
		"issuer":     func(p *idtoken.Payload) { p.Issuer = "https://attacker.test" },
		"expired":    func(p *idtoken.Payload) { p.Expires = time.Now().Add(-time.Minute).Unix() },
		"unverified": func(p *idtoken.Payload) { p.Claims["email_verified"] = false },
		"subject":    func(p *idtoken.Payload) { p.Subject = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := valid()
			mutate(p)
			if _, err := googlePayloadIdentity(p, "andaria-web-client"); err == nil {
				t.Fatal("unsafe claims accepted")
			}
		})
	}
	if _, err := googleTokenIdentity(context.Background(), "not-a-signed-token", "andaria-web-client"); err == nil {
		t.Fatal("invalid signature accepted")
	}
}

func TestGoogleMobileUsesExistingIdentityAndRejectsOtherRoles(t *testing.T) {
	db := testutil.Database(t)
	cfg := config.Config{GoogleClientID: "web-client", RegistrationEnabled: true, PasswordBcryptCost: 10, SessionTTL: time.Hour, SessionIdleTimeout: time.Hour, SessionSecret: strings.Repeat("s", 32)}
	h, err := NewAuthHandler(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	identity := googleIdentity{Subject: "verified-subject", Email: "ana@mobile.test", EmailVerified: true, GivenName: "Ana", FamilyName: "Paz"}
	h.googleMobileVerifier = func(ctx context.Context, token, audience string) (googleIdentity, error) { return identity, nil }
	send := func() *httptest.ResponseRecorder {
		raw, _ := json.Marshal(map[string]string{"id_token": "signed-token-placeholder"})
		r := httptest.NewRequest("POST", "/auth/google/mobile", bytes.NewReader(raw))
		w := httptest.NewRecorder()
		h.GoogleMobileLogin(w, r)
		return w
	}
	var id uint64
	for i := 0; i < 2; i++ {
		w := send()
		if w.Code != 200 || len(w.Result().Cookies()) != 1 {
			t.Fatalf("login got %d %s", w.Code, w.Body.String())
		}
		var envelope struct {
			Data struct {
				User models.PublicUser `json:"user"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Data.User.Role != models.RoleUser {
			t.Fatal("wrong role")
		}
		if id != 0 && id != envelope.Data.User.ID {
			t.Fatal("duplicate identity")
		}
		id = envelope.Data.User.ID
	}
	if err := db.Model(&models.User{}).Where("id = ?", id).Update("role", models.RoleAgency).Error; err != nil {
		t.Fatal(err)
	}
	w := send()
	if w.Code != 403 || len(w.Result().Cookies()) != 0 {
		t.Fatalf("non-tourist received session: %d", w.Code)
	}
}
