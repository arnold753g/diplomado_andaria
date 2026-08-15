package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
	"testing"
)

func TestModuleOnePermissionsAndSessions(t *testing.T) {
	db := testutil.Database(t)
	cfg := testConfig()
	cfg.RegistrationEnabled = true
	cfg.AuthRatePerMinute = 1000
	cfg.APIRatePerMinute = 10000
	cfg.LoginMaxAttempts = 5
	handler, err := New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	password := "integration-password-123"
	hash, err := security.HashPassword(password, 10)
	if err != nil {
		t.Fatal(err)
	}
	admin := models.User{Email: "admin@example.test", PasswordHash: hash, FirstName: "Admin", LastName: "Andaria", Role: models.RoleAdmin, Status: models.StatusActive}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	request := func(method, path string, body any, cookie *http.Cookie, csrf string, want int) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(raw))
		r.Header.Set("Origin", "http://localhost:3000")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		if csrf != "" {
			r.Header.Set("X-CSRF-Token", csrf)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
	login := func(email string) (*http.Cookie, string) {
		t.Helper()
		w := request("POST", "/api/v1/auth/login", map[string]string{"email": email, "password": password}, nil, "", 200)
		var result struct {
			Data struct {
				CSRF string `json:"csrf_token"`
			}
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return w.Result().Cookies()[0], result.Data.CSRF
	}
	payload := map[string]string{"email": "tourist@example.test", "password": password, "first_name": "Turista", "last_name": "Prueba"}
	payload["role"] = "admin"
	request("POST", "/api/v1/auth/register", payload, nil, "", 400)
	delete(payload, "role")
	request("POST", "/api/v1/auth/register", payload, nil, "", 201)
	request("POST", "/api/v1/auth/register", payload, nil, "", 409)
	var tourist models.User
	db.Where("email = ?", payload["email"]).First(&tourist)
	if tourist.Role != models.RoleUser {
		t.Fatal("registration escalated privileges")
	}
	tc, ts := login(tourist.Email)
	request("GET", "/api/v1/admin/users", nil, tc, ts, 403)
	request("POST", "/api/v1/admin/users", payload, tc, ts, 403)
	request("PATCH", "/api/v1/me", map[string]string{"first_name": "Turista", "last_name": "Editado"}, tc, "", 403)
	request("PATCH", "/api/v1/me", map[string]string{"first_name": "Turista", "last_name": "Editado", "phone": "70000000", "document_number": "DEMO", "nationality": "Boliviana"}, tc, ts, 200)
	request("PATCH", "/api/v1/me", map[string]string{"first_name": "Turista", "last_name": "Editado", "role": "admin"}, tc, ts, 400)
	ac, as := login(admin.Email)
	request("PATCH", fmt.Sprintf("/api/v1/admin/users/%d/role", admin.ID), map[string]string{"role": models.RoleUser}, ac, as, 409)
	for i, role := range []string{models.RoleAgency, models.RoleAttraction} {
		email := fmt.Sprintf("manager%d@example.test", i)
		request("POST", "/api/v1/admin/users", map[string]string{"email": email, "password": password, "first_name": "Encargado", "last_name": "Prueba", "role": role}, ac, as, 201)
		mc, ms := login(email)
		request("GET", "/api/v1/me", nil, mc, ms, 200)
		request("GET", "/api/v1/admin/users", nil, mc, ms, 403)
	}
	// Atomic edit must revoke the previous session when the role changes.
	request("PATCH", fmt.Sprintf("/api/v1/admin/users/%d", tourist.ID), map[string]string{"first_name": "Turista", "last_name": "Editado", "role": models.RoleAgency, "status": "active"}, ac, as, 200)
	request("GET", "/api/v1/me", nil, tc, ts, 401)
	tc, ts = login(tourist.Email)
	request("PATCH", fmt.Sprintf("/api/v1/admin/users/%d/status", tourist.ID), map[string]string{"status": "inactive"}, ac, as, 200)
	request("GET", "/api/v1/me", nil, tc, ts, 401)
	request("POST", "/api/v1/auth/login", map[string]string{"email": tourist.Email, "password": password}, nil, "", 403)
	request("POST", "/api/v1/me/password", map[string]string{"current_password": "incorrect", "new_password": "new-password-123"}, ac, as, 422)
	request("GET", "/api/v1/me", nil, ac, as, 200)
	request("POST", "/api/v1/me/password", map[string]string{"current_password": password, "new_password": "new-password-123"}, ac, as, 200)
	request("GET", "/api/v1/me", nil, ac, as, 401)
}
