package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
)

func TestAgencyPaymentsGroupingRosterAuthorizationAndIsolation(t *testing.T) {
	db := testutil.Database(t)
	cfg := testConfig()
	cfg.APIRatePerMinute = 10000
	handler, err := New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	users := []models.User{{Role: models.RoleAgency}, {Role: models.RoleAgency}, {Role: models.RoleUser}, {Role: models.RoleAdmin}, {Role: models.RoleAttraction}}
	tokens := make([]string, len(users))
	csrf := make([]string, len(users))
	now := time.Now().UTC()
	for i := range users {
		users[i].Email = fmt.Sprintf("payments-%d@example.test", i)
		users[i].FirstName = "Prueba"
		users[i].LastName = "Pagos"
		users[i].Status = models.StatusActive
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatal(err)
		}
		tokens[i], _ = security.RandomToken()
		csrf[i], _ = security.CSRFToken(cfg.SessionSecret, tokens[i])
		session := models.Session{ID: fmt.Sprintf("pagos-%d", i), UserID: users[i].ID, TokenHash: security.HashToken(tokens[i]), LastActivityAt: now, ExpiresAt: now.Add(time.Hour)}
		if err := db.Create(&session).Error; err != nil {
			t.Fatal(err)
		}
	}
	agencies := []models.Agency{}
	for i := 0; i < 2; i++ {
		a := models.Agency{Name: fmt.Sprintf("Agencia %d", i), Department: "Tarija", City: "Tarija", Address: "Centro", Phone: "000000", Email: fmt.Sprintf("agencia-%d@example.test", i), ManagerID: users[i].ID, Status: models.StatusActive, Version: 1}
		if err := db.Create(&a).Error; err != nil {
			t.Fatal(err)
		}
		agencies = append(agencies, a)
	}
	packages := []models.TourPackage{}
	for i := 0; i < 3; i++ {
		agency := agencies[0].ID
		if i == 2 {
			agency = agencies[1].ID
		}
		p := models.TourPackage{AgencyID: agency, Name: fmt.Sprintf("Paquete %d", i), DurationDays: 1, CancellationAllowed: true, CancellationNoticeHours: 24, Version: 1, Includes: []string{}, Excludes: []string{}, Bring: []string{}}
		if err := db.Create(&p).Error; err != nil {
			t.Fatal(err)
		}
		packages = append(packages, p)
	}
	departures := []models.TourPackageDeparture{}
	for i, pkg := range []uint64{packages[0].ID, packages[0].ID, packages[1].ID, packages[2].ID} {
		d := models.TourPackageDeparture{PackageID: pkg, StartsAt: now.Add(time.Duration(10+i) * 24 * time.Hour), MeetingAt: now.Add(time.Duration(10+i)*24*time.Hour - time.Hour), BookingOpensAt: now.Add(-time.Hour), BookingClosesAt: now.Add(time.Duration(9+i) * 24 * time.Hour), MinCapacity: 1, MaxCapacity: 100, Status: "open", Version: 1}
		if i == 0 {
			d.MinCapacity = 25
			d.ConfirmedCapacity = 25
			d.HeldCapacity = 2
			d.Status = "confirmed"
		}
		if err := db.Create(&d).Error; err != nil {
			t.Fatal(err)
		}
		departures = append(departures, d)
	}
	var first models.TourPackagePurchase
	for i := 0; i < 28; i++ {
		p := models.TourPackagePurchase{Reference: fmt.Sprintf("PAY-%03d", i), TouristID: users[2].ID, AgencyID: agencies[0].ID, PackageID: packages[0].ID, DepartureID: departures[0].ID, Status: models.PurchaseConfirmed, PaymentMethod: "qr", NationalAdults: 1, CapacityCount: 1, NationalUnitPriceCents: 10000, TotalCents: 10000, TouristName: fmt.Sprintf("Comprador %03d", i), TouristEmail: "comprador@example.test", TouristPhone: "000000", PaymentProof: []byte{1, 2}, ReviewedAt: &now, Version: 1, Minors: []models.PurchaseMinor{}}
		if i == 0 {
			p.Minors = []models.PurchaseMinor{{Age: 3}}
			p.FreeMinorCount = 1
			p.TouristName = "<script>alert('prueba')</script>"
		}
		if i >= 25 {
			p.Status = models.PurchasePaymentReview
			p.ReviewedAt = nil
		}
		if i == 27 {
			p.Status = models.PurchaseRefundPending
			requested, due := now.Add(-96*time.Hour), now.Add(-24*time.Hour)
			p.RefundRequestedAt = &requested
			p.RefundDueAt = &due
			p.RefundReason = "Devolucion de prueba"
		}
		if err := db.Create(&p).Error; err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = p
		}
	}
	rawRequest := func(method, path string, body any, account int) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Origin", "http://localhost:3000")
		if account >= 0 {
			req.AddCookie(&http.Cookie{Name: security.CookieName, Value: tokens[account]})
			req.Header.Set("X-CSRF-Token", csrf[account])
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	request := func(method, path string, body any, account, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := rawRequest(method, path, body, account)
		if w.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
	root := "/api/v1/agency/payments/departures"
	detail := fmt.Sprintf("%s/%d", root, departures[0].ID)
	request("GET", root, nil, -1, 401)
	for _, account := range []int{2, 3, 4} {
		request("GET", root, nil, account, 403)
		request("GET", detail+"/roster", nil, account, 403)
		request("POST", detail+"/authorize", map[string]int{"version": 1}, account, 403)
	}
	request("GET", detail, nil, 1, 404)
	request("GET", detail+"/roster", nil, 1, 404)
	request("POST", detail+"/authorize", map[string]int{"version": 1}, 1, 404)
	var overview struct {
		Data struct {
			Departures []struct {
				ID                 uint64
				PackageID          uint64 `json:"package_id"`
				AuthorizationState string `json:"authorization_state"`
				ConfirmedTravelers int    `json:"confirmed_travelers"`
				CanAuthorize       bool   `json:"can_authorize"`
			}
			Totals struct {
				Departures      int
				PaymentsPending int `json:"payments_pending"`
				RefundsPending  int `json:"refunds_pending"`
				RefundsOverdue  int `json:"refunds_overdue"`
				ReadyDepartures int `json:"ready_departures"`
				ConfirmedAmount int `json:"confirmed_amount_cents"`
			}
			Packages   []struct{ ID uint64 }
			Pagination struct{ Total int }
		}
	}
	w := request("GET", root+"?limit=1", nil, 0, 200)
	if err := json.Unmarshal(w.Body.Bytes(), &overview); err != nil {
		t.Fatal(err)
	}
	if len(overview.Data.Departures) != 1 || overview.Data.Pagination.Total != 3 || overview.Data.Totals.Departures != 3 || overview.Data.Totals.PaymentsPending != 2 || overview.Data.Totals.RefundsPending != 1 || overview.Data.Totals.RefundsOverdue != 1 || overview.Data.Totals.ReadyDepartures != 1 || overview.Data.Totals.ConfirmedAmount != 250000 || len(overview.Data.Packages) != 2 {
		t.Fatalf("incomplete or foreign totals: %s", w.Body.String())
	}
	if row := overview.Data.Departures[0]; row.ID != departures[0].ID || !row.CanAuthorize || row.AuthorizationState != "ready" || row.ConfirmedTravelers != 26 {
		t.Fatalf("manual authorization or traveler count incorrect: %+v", row)
	}
	w = request("GET", root+"?page=2&limit=1", nil, 0, 200)
	json.Unmarshal(w.Body.Bytes(), &overview)
	if len(overview.Data.Departures) != 1 || overview.Data.Departures[0].ID == departures[0].ID {
		t.Fatal("departure pagination repeated first row")
	}
	for _, query := range []string{"?state=invalid", "?package_id=invalid", "?from=2026-99-01", "?from=2026-12-01&to=2026-01-01"} {
		request("GET", root+query, nil, 0, 422)
	}
	date := departures[0].StartsAt.In(time.FixedZone("BO", -4*3600)).Format("2006-01-02")
	w = request("GET", root+"?from="+date+"&to="+date, nil, 0, 200)
	json.Unmarshal(w.Body.Bytes(), &overview)
	if overview.Data.Pagination.Total != 1 {
		t.Fatal("Bolivia date filter excluded selected day")
	}
	request("GET", root+"?state=ready", nil, 0, 200)
	request("GET", root+"?state=attention", nil, 0, 200)
	// Purchasing list paginates within the selected departure and payment status.
	path := fmt.Sprintf("/api/v1/agency/purchases?departure_id=%d&status=confirmed&page=2", departures[0].ID)
	w = request("GET", path, nil, 0, 200)
	var list struct {
		Data struct {
			Purchases  []models.TourPackagePurchase
			Pagination struct{ Total int }
		}
	}
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list.Data.Purchases) != 5 || list.Data.Pagination.Total != 25 {
		t.Fatal("buyer list truncates or ignores departure/status filter")
	}
	w = request("GET", detail+"/roster", nil, 0, 200)
	if !strings.Contains(w.Body.String(), "PAY-024") || strings.Contains(w.Body.String(), "PAY-025") || strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Body.String(), "&lt;script&gt;") || !strings.Contains(w.Header().Get("Cache-Control"), "no-store") {
		t.Fatal("roster is incomplete, unescaped, cached or includes unconfirmed payments")
	}
	// The minimum is commercial; reaching it never authorizes the trip by itself.
	var actual models.TourPackageDeparture
	db.First(&actual, departures[0].ID)
	if actual.AuthorizedAt != nil {
		t.Fatal("trip automatically authorized")
	}
	request("POST", fmt.Sprintf("%s/%d/authorize", root, departures[1].ID), map[string]int{"version": 1}, 0, 422)
	results := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() { results <- rawRequest("POST", detail+"/authorize", map[string]int{"version": 1}, 0).Code }()
	}
	a, b := <-results, <-results
	if !((a == 200 && b == 409) || (a == 409 && b == 200)) {
		t.Fatalf("duplicate authorization accepted or failed: %d %d", a, b)
	}
	db.First(&actual, departures[0].ID)
	if actual.AuthorizedAt == nil || actual.AuthorizedByUserID == nil || *actual.AuthorizedByUserID != users[0].ID {
		t.Fatal("authorization responsibility not recorded")
	}
	request("POST", detail+"/authorize", map[string]int{"version": actual.Version}, 0, 422)
	// A real buyer cancellation loses the minimum and revokes the authorization,
	// but does not cancel the entire departure or automatically refund other buyers.
	cancel := map[string]any{"version": first.Version, "reason": "Cambio de planes de prueba", "refund_method": "bank_transfer", "refund_bank_name": "Banco", "refund_account_holder": "Titular Prueba", "refund_account_number": "123456"}
	request("POST", fmt.Sprintf("/api/v1/me/purchases/%d/cancel", first.ID), cancel, 2, 200)
	actual = models.TourPackageDeparture{}
	db.First(&actual, departures[0].ID)
	if actual.AuthorizationRevokedAt == nil || actual.AuthorizedAt == nil || actual.ConfirmedCapacity != 24 || actual.Status == "cancelled" {
		t.Fatalf("minimum loss did not revoke authorization safely: %+v", actual)
	}
	request("POST", detail+"/authorize", map[string]int{"version": actual.Version}, 0, 422)
	// Restoring the minimum requires another explicit decision, not reactivation.
	if err := db.Model(&actual).Updates(map[string]any{"cupos_confirmados": 25, "version": actual.Version + 1}).Error; err != nil {
		t.Fatal(err)
	}
	actual = models.TourPackageDeparture{}
	db.First(&actual, departures[0].ID)
	if actual.AuthorizationRevokedAt == nil {
		t.Fatal("authorization silently reactivated")
	}
	request("POST", detail+"/authorize", map[string]int{"version": actual.Version}, 0, 200)
	actual = models.TourPackageDeparture{}
	db.First(&actual, departures[0].ID)
	if actual.AuthorizationRevokedAt != nil || actual.AuthorizedAt == nil {
		t.Fatal("explicit reauthorization failed")
	}

	// The main view excludes future unapproved and past departures, while global
	// alerts keep refund obligations visible regardless of the selected view.
	type overviewEnvelope struct {
		Data struct {
			Departures []struct {
				ID                 uint64
				CanCancel          bool   `json:"can_cancel"`
				AuthorizationState string `json:"authorization_state"`
			}
			Pagination struct{ Total int }
			Summary    struct {
				RefundsPending int `json:"refunds_pending"`
				Authorized     int `json:"authorized_departures"`
				Pending        int `json:"pending_departures"`
				Refunds        int `json:"refund_departures"`
				History        int `json:"history_departures"`
			}
		}
	}
	readView := func(state string, account int) overviewEnvelope {
		t.Helper()
		w := request("GET", root+"?state="+state, nil, account, 200)
		var envelope overviewEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return envelope
	}
	past := models.TourPackageDeparture{PackageID: packages[0].ID, StartsAt: now.Add(-48 * time.Hour), MeetingAt: now.Add(-49 * time.Hour), BookingOpensAt: now.Add(-120 * time.Hour), BookingClosesAt: now.Add(-72 * time.Hour), MinCapacity: 1, MaxCapacity: 10, Status: "open", Version: 1}
	if err := db.Create(&past).Error; err != nil {
		t.Fatal(err)
	}
	enabled := readView("authorized", 0)
	if enabled.Data.Pagination.Total != 1 || len(enabled.Data.Departures) != 1 || enabled.Data.Departures[0].ID != actual.ID || !enabled.Data.Departures[0].CanCancel || enabled.Data.Summary.Pending != 2 || enabled.Data.Summary.History != 1 {
		t.Fatalf("main view or navigation counts invalid: %+v", enabled)
	}
	if pending := readView("pending", 0); pending.Data.Pagination.Total != 2 {
		t.Fatal("pending view mixes past or authorized departures")
	}
	if history := readView("history", 0); history.Data.Pagination.Total != 1 || history.Data.Departures[0].CanCancel {
		t.Fatal("history or past cancellation capability invalid")
	}
	cancelPath := fmt.Sprintf("/api/v1/agency/packages/%d/departures/%d/cancel", packages[0].ID, actual.ID)
	request("POST", cancelPath, map[string]any{"version": actual.Version, "reason": "Motivos externos: Corte de carretera"}, 1, 404)
	request("POST", cancelPath, map[string]any{"version": actual.Version, "reason": "abc"}, 0, 422)
	request("POST", fmt.Sprintf("/api/v1/agency/packages/%d/departures/%d/cancel", packages[0].ID, past.ID), map[string]any{"version": 1, "reason": "Motivos externos: Prueba"}, 0, 422)
	var alreadyPending models.TourPackagePurchase
	if err := db.First(&alreadyPending, first.ID).Error; err != nil {
		t.Fatal(err)
	}
	cancellationResults := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() {
			cancellationResults <- rawRequest("POST", cancelPath, map[string]any{"version": actual.Version, "reason": "Motivos externos: Corte de carretera"}, 0).Code
		}()
	}
	x, y := <-cancellationResults, <-cancellationResults
	if !((x == 200 && y == 409) || (x == 409 && y == 200)) {
		t.Fatalf("cancellation was duplicated or failed: %d %d", x, y)
	}
	cancelledDeparture := models.TourPackageDeparture{}
	if err := db.First(&cancelledDeparture, actual.ID).Error; err != nil {
		t.Fatal(err)
	}
	if cancelledDeparture.Status != "cancelled" || cancelledDeparture.AuthorizationRevokedAt == nil || cancelledDeparture.HeldCapacity != 0 || cancelledDeparture.ConfirmedCapacity != 0 || cancelledDeparture.ModifiedByUserID == nil || *cancelledDeparture.ModifiedByUserID != users[0].ID || !strings.Contains(cancelledDeparture.CancellationReason, "Corte de carretera") {
		t.Fatal("authorized departure was not cancelled atomically")
	}
	affected := []models.TourPackagePurchase{}
	if err := db.Where("salida_id=?", actual.ID).Find(&affected).Error; err != nil {
		t.Fatal(err)
	}
	if len(affected) != 28 {
		t.Fatal("cancelled purchases lost")
	}
	for _, purchase := range affected {
		if purchase.Status != models.PurchaseRefundPending || purchase.RefundDueAt == nil {
			t.Fatal("affected payment not moved to refund queue")
		}
		if purchase.ID == first.ID {
			if !purchase.RefundDueAt.Equal(*alreadyPending.RefundDueAt) {
				t.Fatal("existing refund deadline was extended")
			}
		} else if purchase.Reference != "PAY-027" {
			if !strings.Contains(purchase.RefundReason, "Corte de carretera") || purchase.RefundDueAt.Before(time.Now().Add(71*time.Hour)) {
				t.Fatal("new refund reason or deadline missing")
			}
		}
	}
	emptyEnabled := readView("authorized", 0)
	if emptyEnabled.Data.Pagination.Total != 0 || emptyEnabled.Data.Summary.RefundsPending != 28 || emptyEnabled.Data.Summary.Refunds != 1 {
		t.Fatal("refund obligations disappeared with main-view filtering")
	}
	if refunds := readView("refunds", 0); refunds.Data.Pagination.Total != 1 || refunds.Data.Departures[0].CanCancel {
		t.Fatal("cancelled departure not available for refunds")
	}
	if history := readView("history", 0); history.Data.Pagination.Total != 2 {
		t.Fatal("cancelled future departure missing from history")
	}
	if foreign := readView("refunds", 1); foreign.Data.Pagination.Total != 0 || foreign.Data.Summary.RefundsPending != 0 {
		t.Fatal("refund summary leaked across agencies")
	}
	// The tourist-facing contract already exposes the refund state and destination.
	w = request("GET", fmt.Sprintf("/api/v1/me/purchases/%d", affected[1].ID), nil, 2, 200)
	if !strings.Contains(w.Body.String(), "refund_pending") || !strings.Contains(w.Body.String(), "refund_due_at") {
		t.Fatal("tourist cannot see refund status")
	}
	// Cancellation for missing capacity also works without any purchases to refund.
	request("POST", fmt.Sprintf("/api/v1/agency/packages/%d/departures/%d/cancel", packages[0].ID, departures[1].ID), map[string]any{"version": 1, "reason": "Falta de cupos: No se alcanzo el minimo"}, 0, 200)
	if queue := readView("refunds", 0); queue.Data.Summary.RefundsPending != 28 {
		t.Fatal("empty departure cancellation created a refund")
	}
	// Agency deactivation must block new authorizations.
	if err := db.Model(&agencies[0]).Update("estado", models.StatusInactive).Error; err != nil {
		t.Fatal(err)
	}
	request("POST", fmt.Sprintf("%s/%d/authorize", root, departures[1].ID), map[string]int{"version": 1}, 0, 403)
}
