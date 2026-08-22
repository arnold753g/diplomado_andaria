package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
)

func TestAgencyOwnershipValidationAndAssignments(t *testing.T) {
	db := testutil.Database(t)
	cfg := testConfig()
	cfg.AuthRatePerMinute = 1000
	cfg.APIRatePerMinute = 10000
	handler, err := New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	hash, err := security.HashPassword("agency-tests-password", 10)
	if err != nil {
		t.Fatal(err)
	}
	accounts := []models.User{
		{Email: "admin@agency.test", Role: models.RoleAdmin},
		{Email: "manager1@agency.test", Role: models.RoleAgency},
		{Email: "manager2@agency.test", Role: models.RoleAgency},
		{Email: "tourist@agency.test", Role: models.RoleUser},
		{Email: "attraction@agency.test", Role: models.RoleAttraction},
		{Email: "manager3@agency.test", Role: models.RoleAgency},
	}
	for i := range accounts {
		accounts[i].FirstName = "Prueba"
		accounts[i].LastName = "Agencia"
		accounts[i].PasswordHash = hash
		accounts[i].Status = models.StatusActive
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	cookies := make([]*http.Cookie, len(accounts))
	csrf := make([]string, len(accounts))
	rawRequest := func(method, path string, body any, account int, withCSRF bool) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(raw))
		r.Header.Set("Origin", "http://localhost:3000")
		if account >= 0 && cookies[account] != nil {
			r.AddCookie(cookies[account])
			if withCSRF {
				r.Header.Set("X-CSRF-Token", csrf[account])
			}
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	request := func(method, path string, body any, account, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := rawRequest(method, path, body, account, true)
		if w.Code != want {
			t.Fatalf("%s %s got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
	for i, account := range accounts {
		w := request("POST", "/api/v1/auth/login", map[string]string{"email": account.Email, "password": "agency-tests-password"}, -1, 200)
		cookies[i] = w.Result().Cookies()[0]
		var response struct {
			Data struct {
				CSRF string `json:"csrf_token"`
			}
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		csrf[i] = response.Data.CSRF
	}
	read := func(w *httptest.ResponseRecorder) models.Agency {
		t.Helper()
		var response struct{ Data models.Agency }
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		return response.Data
	}
	payload := map[string]any{"name": "Agencia Tarija", "department": "Tarija", "city": "Tarija", "address": "Calle de prueba 123", "phone": "70000000", "email": "contact@agency.test", "manager_id": accounts[1].ID}
	request("GET", "/api/v1/agency", nil, -1, 401)
	request("GET", "/api/v1/agency", nil, 1, 404)
	request("GET", "/api/v1/agency", nil, 3, 403)
	request("GET", "/api/v1/agency", nil, 4, 403)
	request("POST", "/api/v1/admin/agencies", payload, 1, 403)
	if w := rawRequest("POST", "/api/v1/admin/agencies", payload, 0, false); w.Code != 403 {
		t.Fatal("missing CSRF accepted")
	}
	invalid := maps.Clone(payload)
	invalid["manager_id"] = accounts[3].ID
	request("POST", "/api/v1/admin/agencies", invalid, 0, 422)
	agency := read(request("POST", "/api/v1/admin/agencies", payload, 0, 201))
	if agency.MinimumPayingAge != 6 || agency.ManagerID != accounts[1].ID || agency.Version != 1 {
		t.Fatalf("wrong defaults: %+v", agency)
	}
	request("POST", "/api/v1/admin/agencies", payload, 0, 409)
	path := fmt.Sprintf("/api/v1/admin/agencies/%d", agency.ID)
	request("GET", path, nil, 1, 403)
	request("GET", "/api/v1/agency", nil, 2, 404)
	own := maps.Clone(payload)
	delete(own, "manager_id")
	own["version"] = 1
	own["minimum_paying_age"] = 8
	own["name"] = "Agencia editada"
	forged := maps.Clone(own)
	forged["manager_id"] = accounts[2].ID
	request("PUT", "/api/v1/agency", forged, 1, 403)
	forged = maps.Clone(own)
	forged["status"] = "inactive"
	request("PUT", "/api/v1/agency", forged, 1, 403)
	forged = maps.Clone(own)
	forged["minimum_paying_age"] = -1
	request("PUT", "/api/v1/agency", forged, 1, 422)
	forged = maps.Clone(own)
	forged["accepts_transfer"] = true
	request("PUT", "/api/v1/agency", forged, 1, 422)
	forged = maps.Clone(own)
	forged["accepts_qr"] = true
	forged["qr_image"] = "data:image/svg+xml;base64,PHN2Zy8+"
	request("PUT", "/api/v1/agency", forged, 1, 422)
	updated := read(request("PUT", "/api/v1/agency", own, 1, 200))
	if updated.Name != "Agencia editada" || updated.MinimumPayingAge != 8 || updated.Version != 2 {
		t.Fatal("agency edits not persisted")
	}
	request("PUT", "/api/v1/agency", own, 1, 409)
	// A canonical QR and bank details persist in the same atomic edit.
	var pngBytes bytes.Buffer
	if err := png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	own["version"] = 2
	own["accepts_qr"] = true
	own["qr_image"] = "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes.Bytes())
	own["accepts_transfer"] = true
	own["bank_name"] = "Banco de prueba"
	own["account_holder"] = "Titular de prueba"
	own["account_number"] = "00001234"
	request("PUT", "/api/v1/agency", own, 1, 200)
	var saved models.Agency
	if err := db.First(&saved, agency.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(saved.QRImage) == 0 || saved.AccountNumber != "00001234" {
		t.Fatal("payment configuration did not persist")
	}
	request("PATCH", fmt.Sprintf("/api/v1/admin/users/%d/role", accounts[1].ID), map[string]string{"role": models.RoleUser}, 0, 409)
	request("PATCH", fmt.Sprintf("/api/v1/admin/users/%d", accounts[1].ID), map[string]string{"first_name": "Prueba", "last_name": "Agencia", "role": models.RoleUser, "status": "active"}, 0, 409)
	// Reassignment immediately changes ownership without relying on browser navigation.
	adminEdit := maps.Clone(own)
	adminEdit["version"] = 3
	adminEdit["manager_id"] = accounts[2].ID
	request("PUT", path, adminEdit, 0, 200)
	request("GET", "/api/v1/agency", nil, 1, 404)
	request("PUT", "/api/v1/agency", own, 1, 404)
	request("GET", "/api/v1/agency", nil, 2, 200)
	adminEdit["version"] = 4
	adminEdit["status"] = "inactive"
	adminEdit["published"] = true
	inactive := read(request("PUT", path, adminEdit, 0, 200))
	if inactive.Published || inactive.Status != "inactive" {
		t.Fatal("inactive agency remained published")
	}
	own["version"] = 5
	request("PUT", "/api/v1/agency", own, 2, 403)
	list := request("GET", "/api/v1/admin/agencies?search=editada&status=inactive", nil, 0, 200)
	if !bytes.Contains(list.Body.Bytes(), []byte("Agencia editada")) {
		t.Fatal("search/filter failed")
	}
	if bytes.Contains(list.Body.Bytes(), []byte("data:image")) {
		t.Fatal("list included QR image")
	}
	options := request("GET", "/api/v1/admin/agencies/managers", nil, 0, 200)
	if bytes.Contains(options.Body.Bytes(), []byte(accounts[2].Email)) {
		t.Fatal("assigned manager offered again")
	}
	// Database uniqueness and application locking protect simultaneous assignments.
	concurrent := maps.Clone(payload)
	concurrent["manager_id"] = accounts[5].ID
	results := make(chan int, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- rawRequest("POST", "/api/v1/admin/agencies", concurrent, 0, true).Code
		}()
	}
	wg.Wait()
	close(results)
	created, conflict := 0, 0
	for code := range results {
		if code == 201 {
			created++
		} else if code == 409 {
			conflict++
		} else {
			t.Fatalf("unexpected concurrent result %d", code)
		}
	}
	if created != 1 || conflict != 1 {
		t.Fatalf("created=%d conflicts=%d", created, conflict)
	}
}
