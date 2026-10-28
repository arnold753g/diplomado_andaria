package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
	"starter-backend/internal/config"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
)

func securityAuth(t *testing.T) (*AuthHandler, models.User) {
	t.Helper()
	db := testutil.Database(t)
	cfg := config.Config{PasswordBcryptCost: 10, LoginMaxAttempts: 5, LoginLockout: time.Minute,
		SessionSecret: strings.Repeat("a", 32), SessionTTL: time.Hour, SessionIdleTimeout: time.Minute, RegistrationEnabled: true}
	h, err := NewAuthHandler(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := security.HashPassword("security-test-password", 10)
	if err != nil {
		t.Fatal(err)
	}
	user := models.User{Email: "security@example.test", PasswordHash: hash, FirstName: "Prueba", LastName: "Seguridad", Role: models.RoleUser, Status: models.StatusActive}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	return h, user
}

func securityRequest(handler http.HandlerFunc, body any) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(raw)))
	w := httptest.NewRecorder()
	handler(w, r)
	return w
}

func TestConcurrentFailedLoginsKeepLockoutAndExpire(t *testing.T) {
	h, user := securityAuth(t)
	var group sync.WaitGroup
	codes := make(chan int, 12)
	for i := 0; i < 12; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			w := securityRequest(h.Login, map[string]string{"email": user.Email, "password": "incorrect-password"})
			codes <- w.Code
		}()
	}
	group.Wait()
	close(codes)
	for code := range codes {
		if code != 401 {
			t.Fatalf("failed login status %d", code)
		}
	}
	if err := h.db.First(&user, user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if user.LockedUntil == nil || user.FailedLoginAttempts != 0 {
		t.Fatal("parallel failures lost lockout")
	}
	locked := *user.LockedUntil
	w := securityRequest(h.Login, map[string]string{"email": user.Email, "password": "security-test-password"})
	if w.Code != 429 {
		t.Fatalf("locked account authenticated: %d", w.Code)
	}
	securityRequest(h.Login, map[string]string{"email": user.Email, "password": "incorrect-password"})
	if err := h.db.First(&user, user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !user.LockedUntil.Equal(locked) {
		t.Fatal("failed login extended lockout")
	}
	if err := h.db.Model(&user).Update("bloqueado_hasta", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	w = securityRequest(h.Login, map[string]string{"email": user.Email, "password": "security-test-password"})
	if w.Code != 200 {
		t.Fatalf("expired lockout: %d %s", w.Code, w.Body.String())
	}
}

func TestFailedLoginWriteErrorIsNotReportedAsSuccessfulAccounting(t *testing.T) {
	h, user := securityAuth(t)
	if err := h.db.Callback().Update().Before("gorm:update").Register("fail_login_write", func(tx *gorm.DB) {
		if tx.Statement.Table == "usuarios" {
			tx.AddError(errors.New("injected failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	defer h.db.Callback().Update().Remove("fail_login_write")
	w := securityRequest(h.Login, map[string]string{"email": user.Email, "password": "incorrect-password"})
	if w.Code != 503 || !strings.Contains(w.Body.String(), "AUTH_UNAVAILABLE") {
		t.Fatalf("write failure: %d %s", w.Code, w.Body.String())
	}
}

func TestLogoutRevokesReplayAndReportsWriteFailure(t *testing.T) {
	h, user := securityAuth(t)
	login := securityRequest(h.Login, map[string]string{"email": user.Email, "password": "security-test-password"})
	if login.Code != 200 {
		t.Fatal(login.Body.String())
	}
	cookie := login.Result().Cookies()[0]
	var session models.Session
	if err := h.db.Where("hash_token = ?", security.HashToken(cookie.Value)).First(&session).Error; err != nil {
		t.Fatal(err)
	}
	request := func() *http.Request {
		return middleware.WithPrincipal(httptest.NewRequest("POST", "/", nil), middleware.Principal{User: user, SessionID: session.ID})
	}
	if err := h.db.Callback().Update().Before("gorm:update").Register("fail_revoke", func(tx *gorm.DB) {
		if tx.Statement.Table == "sesiones_autenticacion" {
			tx.AddError(errors.New("injected failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	failed := httptest.NewRecorder()
	h.Logout(failed, request())
	if failed.Code != 503 || len(failed.Result().Cookies()) == 0 || failed.Result().Cookies()[0].MaxAge != -1 {
		t.Fatalf("logout write failure: %d", failed.Code)
	}
	h.db.Callback().Update().Remove("fail_revoke")
	success := httptest.NewRecorder()
	h.Logout(success, request())
	if success.Code != 200 {
		t.Fatal(success.Body.String())
	}
	replay := httptest.NewRequest("GET", "/", nil)
	replay.AddCookie(cookie)
	response := httptest.NewRecorder()
	middleware.NewAuthenticator(h.db, h.cfg).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("revoked token replay accepted") })).ServeHTTP(response, replay)
	if response.Code != 401 {
		t.Fatalf("replay status %d", response.Code)
	}
}

func TestVerifiedRegistrationSingleUseExpiryAndNoEmailReservation(t *testing.T) {
	h, _ := securityAuth(t)
	h.cfg.EmailVerificationRequired = true
	h.cfg.SMTPHost = "smtp.example.test"
	h.cfg.FrontendURL = "https://andaria.example.test"
	var link string
	h.sendVerification = func(_ context.Context, _ string, value string) error { link = value; return nil }
	payload := map[string]string{"email": "new@example.test", "password": "security-test-password", "first_name": "Nuevo", "last_name": "Viajero"}
	registered := securityRequest(h.Register, payload)
	if registered.Code != 202 {
		t.Fatalf("register: %d %s", registered.Code, registered.Body.String())
	}
	parsed, _ := url.Parse(link)
	fragment, _ := url.ParseQuery(parsed.Fragment)
	token := fragment.Get("token")
	if len(token) != 64 || strings.Contains(registered.Body.String(), token) {
		t.Fatal("verification token missing or exposed")
	}
	var count int64
	h.db.Model(&models.User{}).Where("correo = ?", payload["email"]).Count(&count)
	if count != 0 {
		t.Fatal("unverified registration reserved email")
	}
	var pending pendingRegistration
	if err := h.db.Where("correo = ?", payload["email"]).First(&pending).Error; err != nil {
		t.Fatal(err)
	}
	if pending.TokenHash == token || pending.PasswordHash == payload["password"] {
		t.Fatal("plaintext token or password stored")
	}
	if w := securityRequest(h.Register, payload); w.Code != 429 {
		t.Fatal("resend cooldown bypassed")
	}
	if err := h.db.Model(&pending).Update("creado_en", time.Now().Add(-2*time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	if w := securityRequest(h.Register, payload); w.Code != 202 {
		t.Fatalf("resend: %d %s", w.Code, w.Body.String())
	}
	if w := securityRequest(h.VerifyEmail, map[string]string{"token": token}); w.Code != 422 {
		t.Fatal("superseded verification token accepted")
	}
	parsed, _ = url.Parse(link)
	fragment, _ = url.ParseQuery(parsed.Fragment)
	token = fragment.Get("token")
	if w := securityRequest(h.Register, payload); w.Code != 429 {
		t.Fatal("resend did not reset the per-email cooldown")
	}
	// Expired links cannot create accounts, but transient failures can be retried.
	h.db.Model(&pending).Update("vence_en", time.Now().Add(-time.Minute))
	if w := securityRequest(h.VerifyEmail, map[string]string{"token": token}); w.Code != 422 {
		t.Fatal("expired token accepted")
	}
	h.db.Model(&pending).Update("vence_en", time.Now().Add(time.Minute))
	if w := securityRequest(h.VerifyEmail, map[string]string{"token": token}); w.Code != 200 {
		t.Fatalf("verify: %d %s", w.Code, w.Body.String())
	}
	if w := securityRequest(h.VerifyEmail, map[string]string{"token": token}); w.Code != 422 {
		t.Fatal("token replay accepted")
	}
	var user models.User
	h.db.Where("correo = ?", payload["email"]).First(&user)
	if user.Role != models.RoleUser || user.EmailVerifiedAt == nil {
		t.Fatal("verified account invalid")
	}
	if w := securityRequest(h.Login, map[string]string{"email": user.Email, "password": payload["password"]}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	// A pending email does not block its owner registering via verified Google.
	payload["email"] = "google-owner@example.test"
	if w := securityRequest(h.Register, payload); w.Code != 202 {
		t.Fatal(w.Body.String())
	}
	if _, err := h.resolveGoogleUser(googleIdentity{Subject: "verified-owner", Email: payload["email"], EmailVerified: true}, oauthAttempt{}); err != nil {
		t.Fatal(err)
	}
}
