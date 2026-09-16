package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
)

func TestRoleDashboardsExposeScopedOperationalCounts(t *testing.T) {
	db := testutil.Database(t)
	cfg := testConfig()
	cfg.AuthRatePerMinute, cfg.APIRatePerMinute = 1000, 10000
	handler, err := New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := security.HashPassword("dashboard-tests-password", 10)
	accounts := []models.User{
		{Email: "admin@dashboard.test", Role: models.RoleAdmin},
		{Email: "tourist@dashboard.test", Role: models.RoleUser},
		{Email: "agency@dashboard.test", Role: models.RoleAgency},
		{Email: "attraction@dashboard.test", Role: models.RoleAttraction},
	}
	for i := range accounts {
		accounts[i].FirstName, accounts[i].LastName = "Panel", "Prueba"
		accounts[i].PasswordHash, accounts[i].Status = hash, models.StatusActive
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	agency := models.Agency{Name: "Agencia Panel", Department: "Tarija", City: "Tarija", Address: "Centro", Phone: "70000000", Email: "panel@agency.test", ManagerID: accounts[2].ID, Status: models.StatusActive, Published: true, MinimumPayingAge: 6, Version: 1}
	if err := db.Create(&agency).Error; err != nil {
		t.Fatal(err)
	}
	attractions := []models.Attraction{
		{Name: "Atracción publicada", Department: "Tarija", City: "Tarija", Address: "Centro", ManagerID: accounts[3].ID, Status: models.StatusActive, Published: true, Version: 1, ScheduleMode: "unspecified", SeasonMode: "unspecified", OpeningDays: []int{}},
		{Name: "Atracción borrador", Department: "Tarija", City: "Tarija", Address: "Centro", ManagerID: accounts[3].ID, Status: models.StatusActive, Published: false, Version: 1, ScheduleMode: "unspecified", SeasonMode: "unspecified", OpeningDays: []int{}},
	}
	if err := db.Create(&attractions).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.AttractionFavorite{UserID: accounts[1].ID, AttractionID: attractions[0].ID}).Error; err != nil {
		t.Fatal(err)
	}
	item := models.TourPackage{AgencyID: agency.ID, Name: "Paquete panel", Description: "Paquete para probar el panel.", DurationDays: 1, NationalPriceCents: 10000, Published: true, Version: 1, Includes: []string{}, Excludes: []string{}, Bring: []string{}}
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	minimumMarkedAt := now.Add(-time.Hour)
	departures := []models.TourPackageDeparture{
		{PackageID: item.ID, StartsAt: now.Add(72 * time.Hour), MeetingAt: now.Add(71*time.Hour + 30*time.Minute), BookingOpensAt: now.Add(-24 * time.Hour), BookingClosesAt: now.Add(48 * time.Hour), MinCapacity: 1, MaxCapacity: 5, HeldCapacity: 1, Status: "open", Version: 1},
		{PackageID: item.ID, StartsAt: now.Add(96 * time.Hour), MeetingAt: now.Add(95*time.Hour + 30*time.Minute), BookingOpensAt: now.Add(-48 * time.Hour), BookingClosesAt: now.Add(-time.Hour), MinCapacity: 2, MaxCapacity: 5, Status: "minimum_review", MinimumReviewAt: &minimumMarkedAt, Version: 1},
	}
	if err := db.Create(&departures).Error; err != nil {
		t.Fatal(err)
	}
	requestedAt, dueAt := now, now.Add(12*time.Hour)
	purchases := []models.TourPackagePurchase{
		{Reference: "DASH-PAYMENT", TouristID: accounts[1].ID, AgencyID: agency.ID, PackageID: item.ID, DepartureID: departures[0].ID, Status: models.PurchasePaymentReview, PaymentMethod: "transfer", NationalAdults: 1, Minors: []models.PurchaseMinor{}, CapacityCount: 1, NationalUnitPriceCents: 10000, TotalCents: 10000, TouristName: "Panel Prueba", TouristEmail: accounts[1].Email, PaymentProof: []byte{1}, Version: 1},
		{Reference: "DASH-REFUND", TouristID: accounts[1].ID, AgencyID: agency.ID, PackageID: item.ID, DepartureID: departures[1].ID, Status: models.PurchaseRefundPending, PaymentMethod: "transfer", NationalAdults: 1, Minors: []models.PurchaseMinor{}, CapacityCount: 1, NationalUnitPriceCents: 10000, TotalCents: 10000, TouristName: "Panel Prueba", TouristEmail: accounts[1].Email, PaymentProof: []byte{1}, RefundReason: "Salida cancelada para prueba", RefundRequestedAt: &requestedAt, RefundDueAt: &dueAt, Version: 1},
	}
	if err := db.Create(&purchases).Error; err != nil {
		t.Fatal(err)
	}

	cookies := make([]*http.Cookie, len(accounts))
	request := func(path string, account, want int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if account >= 0 && cookies[account] != nil {
			req.AddCookie(cookies[account])
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != want {
			t.Fatalf("GET %s got %d want %d: %s", path, response.Code, want, response.Body.String())
		}
		return response
	}
	for i, account := range accounts {
		raw, _ := json.Marshal(map[string]string{"email": account.Email, "password": "dashboard-tests-password"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(raw))
		req.Header.Set("Origin", "http://localhost:3000")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("login %s: %s", account.Role, response.Body.String())
		}
		cookies[i] = response.Result().Cookies()[0]
	}

	request("/api/v1/me/dashboard", -1, http.StatusUnauthorized)
	request("/api/v1/me/dashboard", 0, http.StatusForbidden)

	var tourist struct {
		Data struct {
			Role    string `json:"role"`
			Tourist struct {
				Favorites       int64 `json:"favorites"`
				ActivePurchases int64 `json:"active_purchases"`
				PendingRefunds  int64 `json:"pending_refunds"`
				NextPurchase    *struct {
					Reference string `json:"reference"`
				} `json:"next_purchase"`
			} `json:"tourist"`
		} `json:"data"`
	}
	if err := json.Unmarshal(request("/api/v1/me/dashboard", 1, http.StatusOK).Body.Bytes(), &tourist); err != nil {
		t.Fatal(err)
	}
	if tourist.Data.Role != models.RoleUser || tourist.Data.Tourist.Favorites != 1 || tourist.Data.Tourist.ActivePurchases != 1 || tourist.Data.Tourist.PendingRefunds != 1 || tourist.Data.Tourist.NextPurchase == nil {
		t.Fatalf("unexpected tourist dashboard: %+v", tourist.Data)
	}

	var agencyResult struct {
		Data struct {
			Agency struct {
				Assigned         bool  `json:"assigned"`
				PaymentsToReview int64 `json:"payments_to_review"`
				MinimumReviews   int64 `json:"minimum_reviews"`
				PendingRefunds   int64 `json:"pending_refunds"`
				RefundsDueSoon   int64 `json:"refunds_due_soon"`
			} `json:"agency"`
		} `json:"data"`
	}
	if err := json.Unmarshal(request("/api/v1/me/dashboard", 2, http.StatusOK).Body.Bytes(), &agencyResult); err != nil {
		t.Fatal(err)
	}
	if !agencyResult.Data.Agency.Assigned || agencyResult.Data.Agency.PaymentsToReview != 1 || agencyResult.Data.Agency.MinimumReviews != 1 || agencyResult.Data.Agency.PendingRefunds != 1 || agencyResult.Data.Agency.RefundsDueSoon != 1 {
		t.Fatalf("unexpected agency dashboard: %+v", agencyResult.Data.Agency)
	}

	var attractionResult struct {
		Data struct {
			Manager struct {
				AssignedAttractions  int64 `json:"assigned_attractions"`
				PublishedAttractions int64 `json:"published_attractions"`
				DraftAttractions     int64 `json:"draft_attractions"`
			} `json:"attraction_manager"`
		} `json:"data"`
	}
	if err := json.Unmarshal(request("/api/v1/me/dashboard", 3, http.StatusOK).Body.Bytes(), &attractionResult); err != nil {
		t.Fatal(err)
	}
	if attractionResult.Data.Manager.AssignedAttractions != 2 || attractionResult.Data.Manager.PublishedAttractions != 1 || attractionResult.Data.Manager.DraftAttractions != 1 {
		t.Fatalf("unexpected attraction dashboard: %+v", attractionResult.Data.Manager)
	}

	var adminResult struct {
		Data map[string]int64 `json:"data"`
	}
	if err := json.Unmarshal(request("/api/v1/admin/dashboard", 0, http.StatusOK).Body.Bytes(), &adminResult); err != nil {
		t.Fatal(err)
	}
	if adminResult.Data["users"] != 4 || adminResult.Data["agencies"] != 1 || adminResult.Data["attractions"] != 2 || adminResult.Data["packages"] != 1 || adminResult.Data["purchases"] != 2 {
		t.Fatalf("unexpected admin dashboard: %+v", adminResult.Data)
	}
}
