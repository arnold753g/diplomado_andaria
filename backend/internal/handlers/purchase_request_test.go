package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestMobilePurchaseIdempotencyAndExpectedPrice(t *testing.T) {
	db := testutil.Database(t)
	hash, err := security.HashPassword("purchase-test-password", 10)
	if err != nil {
		t.Fatal(err)
	}
	accounts := []models.User{
		{Email: "agency@mobile.test", Role: models.RoleAgency, FirstName: "Agencia", LastName: "Prueba", Status: models.StatusActive, PasswordHash: hash},
		{Email: "tourist@mobile.test", Role: models.RoleUser, FirstName: "Turista", LastName: "Prueba", Status: models.StatusActive, PasswordHash: hash},
		{Email: "second@mobile.test", Role: models.RoleUser, FirstName: "Otra", LastName: "Cuenta", Status: models.StatusActive, PasswordHash: hash},
	}
	for i := range accounts {
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	agency := models.Agency{Name: "Agencia móvil", Department: "Tarija", City: "Tarija", Address: "Centro", Phone: "70000000", Email: "ventas@mobile.test", ManagerID: accounts[0].ID, Status: models.StatusActive, Published: true, MinimumPayingAge: 6, AcceptsTransfer: true, BankName: "Banco", AccountHolder: "Agencia", AccountNumber: "1234", Version: 1}
	if err := db.Create(&agency).Error; err != nil {
		t.Fatal(err)
	}
	item := models.TourPackage{AgencyID: agency.ID, Name: "Experiencia móvil", Description: "Descripción de prueba", DurationDays: 1, NationalPriceCents: 25000, ForeignSurchargeCents: 5000, Published: true, Version: 1, Includes: []string{}, Excludes: []string{}, Bring: []string{}}
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	departure := models.TourPackageDeparture{PackageID: item.ID, StartsAt: now.Add(72 * time.Hour), MeetingAt: now.Add(71 * time.Hour), BookingOpensAt: now.Add(-time.Hour), BookingClosesAt: now.Add(48 * time.Hour), MinCapacity: 1, MaxCapacity: 20, Status: "open", Version: 1}
	if err := db.Create(&departure).Error; err != nil {
		t.Fatal(err)
	}
	var imageBuffer bytes.Buffer
	if err := png.Encode(&imageBuffer, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{"departure_id": departure.ID, "payment_method": "transfer", "national_adults": 1, "foreign_adults": 0, "minors": []any{}, "payment_proof": "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageBuffer.Bytes()), "expected_total_cents": 25000}
	handler := NewPurchaseHandler(db)
	send := func(key string, body map[string]any, user models.User) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/v1/me/purchases", bytes.NewReader(raw))
		req.Header.Set("Idempotency-Key", key)
		req = middleware.WithPrincipal(req, middleware.Principal{User: user})
		response := httptest.NewRecorder()
		handler.Create(response, req)
		return response
	}
	const key = "mobile-request-00000001"
	var responses [2]*httptest.ResponseRecorder
	var wg sync.WaitGroup
	for i := range responses {
		wg.Add(1)
		go func(i int) { defer wg.Done(); responses[i] = send(key, payload, accounts[1]) }(i)
	}
	wg.Wait()
	statuses := map[int]int{}
	var purchaseID uint64
	for _, response := range responses {
		statuses[response.Code]++
		var envelope struct {
			Data struct {
				ID uint64 `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Data.ID == 0 {
			t.Fatalf("request failed: %d %s", response.Code, response.Body.String())
		}
		if purchaseID != 0 && purchaseID != envelope.Data.ID {
			t.Fatal("retry created a second purchase")
		}
		purchaseID = envelope.Data.ID
	}
	if statuses[201] != 1 || statuses[200] != 1 {
		t.Fatalf("unexpected statuses: %v", statuses)
	}
	get := func(id uint64, user models.User, serve http.HandlerFunc) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/api/v1/me/resource", nil)
		req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatUint(id, 10)})
		req = middleware.WithPrincipal(req, middleware.Principal{User: user})
		response := httptest.NewRecorder()
		serve(response, req)
		return response
	}
	for _, response := range []*httptest.ResponseRecorder{
		get(item.ID, accounts[1], handler.PaymentOptions),
		get(purchaseID, accounts[1], handler.GetMine),
	} {
		var envelope struct {
			Data struct {
				Phone string `json:"agency_phone"`
				Email string `json:"agency_email"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if response.Code != 200 || envelope.Data.Phone != agency.Phone || envelope.Data.Email != agency.Email {
			t.Fatalf("commercial contact missing: %d %s", response.Code, response.Body.String())
		}
		if bytes.Contains(response.Body.Bytes(), []byte(accounts[0].Email)) {
			t.Fatal("manager's personal email exposed as commercial contact")
		}
	}
	if response := get(purchaseID, accounts[2], handler.GetMine); response.Code != 404 {
		t.Fatalf("another tourist can read the purchase: %d", response.Code)
	}
	assertCapacity := func(want int) {
		t.Helper()
		var current models.TourPackageDeparture
		if err := db.First(&current, departure.ID).Error; err != nil {
			t.Fatal(err)
		}
		if current.HeldCapacity != want {
			t.Fatalf("held capacity got %d want %d", current.HeldCapacity, want)
		}
	}
	assertCapacity(1)
	changed := map[string]any{}
	for k, v := range payload {
		changed[k] = v
	}
	changed["national_adults"] = 2
	changed["expected_total_cents"] = 50000
	if response := send(key, changed, accounts[1]); response.Code != 409 {
		t.Fatalf("key reuse got %d: %s", response.Code, response.Body.String())
	}
	assertCapacity(1)
	if response := send(key, payload, accounts[2]); response.Code != 201 {
		t.Fatalf("separate account got %d: %s", response.Code, response.Body.String())
	}
	assertCapacity(2)
	changed["expected_total_cents"] = 1
	if response := send("mobile-price-change-0001", changed, accounts[1]); response.Code != 409 {
		t.Fatalf("price mismatch got %d", response.Code)
	}
	assertCapacity(2)
	if response := send("invalid", payload, accounts[1]); response.Code != 422 {
		t.Fatalf("invalid key got %d", response.Code)
	}
	// Legacy web requests without the new optional fields keep working.
	delete(changed, "expected_total_cents")
	if response := send("", changed, accounts[1]); response.Code != 201 {
		t.Fatalf("legacy request got %d: %s", response.Code, response.Body.String())
	}
	assertCapacity(4)
	// A replay still returns the accepted purchase after its departure closes.
	if err := db.Model(&departure).Update("status", "closed").Error; err != nil {
		t.Fatal(err)
	}
	if response := send(key, payload, accounts[1]); response.Code != 200 {
		t.Fatalf("late retry got %d: %s", response.Code, response.Body.String())
	}
	assertCapacity(4)
}
