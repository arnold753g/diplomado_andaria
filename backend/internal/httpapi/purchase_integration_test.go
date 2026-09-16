package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/services"
	"starter-backend/internal/testutil"
)

func TestPurchaseCapacityPricingReviewAndRoleIsolation(t *testing.T) {
	db := testutil.Database(t)
	cfg := testConfig()
	cfg.AuthRatePerMinute, cfg.APIRatePerMinute = 1000, 10000
	handler, err := New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := security.HashPassword("purchase-tests-password", 10)
	accounts := []models.User{
		{Email: "agency-one@purchases.test", Role: models.RoleAgency},
		{Email: "agency-two@purchases.test", Role: models.RoleAgency},
		{Email: "tourist-one@purchases.test", Role: models.RoleUser},
		{Email: "tourist-two@purchases.test", Role: models.RoleUser},
		{Email: "admin@purchases.test", Role: models.RoleAdmin},
		{Email: "attraction@purchases.test", Role: models.RoleAttraction},
	}
	for i := range accounts {
		accounts[i].FirstName, accounts[i].LastName, accounts[i].PasswordHash, accounts[i].Status = "Prueba", "Compra", hash, models.StatusActive
		accounts[i].Phone, accounts[i].DocumentNumber, accounts[i].Nationality = "70000000", fmt.Sprintf("DOC-%d", i), "Bolivia"
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	agencies := []models.Agency{
		{Name: "Agencia Comprable", Department: "Tarija", City: "Tarija", Address: "Centro", Phone: "70000001", Email: "uno@purchase.test", ManagerID: accounts[0].ID, Status: models.StatusActive, Published: true, MinimumPayingAge: 6, AcceptsTransfer: true, BankName: "Banco", AccountHolder: "Agencia Comprable", AccountNumber: "123", Version: 1},
		{Name: "Agencia Ajena", Department: "Tarija", City: "Tarija", Address: "Centro", Phone: "70000002", Email: "dos@purchase.test", ManagerID: accounts[1].ID, Status: models.StatusActive, Published: true, MinimumPayingAge: 8, AcceptsTransfer: true, BankName: "Banco", AccountHolder: "Agencia Ajena", AccountNumber: "456", Version: 1},
	}
	for i := range agencies {
		if err := db.Create(&agencies[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	item := models.TourPackage{AgencyID: agencies[0].ID, Name: "Paquete comprable", Description: "Paquete preparado para verificar compras.", DurationDays: 1, NationalPriceCents: 25000, ForeignSurchargeCents: 5000, CancellationAllowed: true, CancellationNoticeHours: 48, Published: true, Version: 1, Includes: []string{}, Excludes: []string{}, Bring: []string{}}
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	departures := []models.TourPackageDeparture{
		{PackageID: item.ID, StartsAt: now.Add(10 * 24 * time.Hour), MeetingAt: now.Add(10*24*time.Hour - 30*time.Minute), BookingOpensAt: now.Add(-24 * time.Hour), BookingClosesAt: now.Add(9 * 24 * time.Hour), MinCapacity: 4, MaxCapacity: 10, Status: "open", Version: 1},
		{PackageID: item.ID, StartsAt: now.Add(11 * 24 * time.Hour), MeetingAt: now.Add(11*24*time.Hour - 30*time.Minute), BookingOpensAt: now.Add(-24 * time.Hour), BookingClosesAt: now.Add(10 * 24 * time.Hour), MinCapacity: 1, MaxCapacity: 1, Status: "open", Version: 1},
		{PackageID: item.ID, StartsAt: now.Add(12 * 24 * time.Hour), MeetingAt: now.Add(12*24*time.Hour - 30*time.Minute), BookingOpensAt: now.Add(-24 * time.Hour), BookingClosesAt: now.Add(11 * 24 * time.Hour), MinCapacity: 2, MaxCapacity: 4, Status: "open", Version: 1},
		{PackageID: item.ID, StartsAt: now.Add(13 * 24 * time.Hour), MeetingAt: now.Add(13*24*time.Hour - 30*time.Minute), BookingOpensAt: now.Add(-24 * time.Hour), BookingClosesAt: now.Add(12 * 24 * time.Hour), MinCapacity: 2, MaxCapacity: 4, Status: "open", Version: 1},
		{PackageID: item.ID, StartsAt: now.Add(14 * 24 * time.Hour), MeetingAt: now.Add(14*24*time.Hour - 30*time.Minute), BookingOpensAt: now.Add(-24 * time.Hour), BookingClosesAt: now.Add(13 * 24 * time.Hour), MinCapacity: 3, MaxCapacity: 4, Status: "open", Version: 1},
		{PackageID: item.ID, StartsAt: now.Add(15 * 24 * time.Hour), MeetingAt: now.Add(15*24*time.Hour - 30*time.Minute), BookingOpensAt: now.Add(-24 * time.Hour), BookingClosesAt: now.Add(14 * 24 * time.Hour), MinCapacity: 1, MaxCapacity: 3, Status: "open", Version: 1},
		{PackageID: item.ID, StartsAt: now.Add(24 * time.Hour), MeetingAt: now.Add(23*time.Hour + 30*time.Minute), BookingOpensAt: now.Add(-24 * time.Hour), BookingClosesAt: now.Add(12 * time.Hour), MinCapacity: 1, MaxCapacity: 3, Status: "open", Version: 1},
	}
	for i := range departures {
		if err := db.Create(&departures[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	cookies := make([]*http.Cookie, len(accounts))
	csrf := make([]string, len(accounts))
	rawRequest := func(method, path string, body any, account int) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Origin", "http://localhost:3000")
		if account >= 0 && cookies[account] != nil {
			req.AddCookie(cookies[account])
			req.Header.Set("X-CSRF-Token", csrf[account])
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	request := func(method, path string, body any, account, want int) *httptest.ResponseRecorder {
		t.Helper()
		response := rawRequest(method, path, body, account)
		if response.Code != want {
			t.Fatalf("%s %s got %d want %d: %s", method, path, response.Code, want, response.Body.String())
		}
		return response
	}
	for i, account := range accounts {
		response := request("POST", "/api/v1/auth/login", map[string]string{"email": account.Email, "password": "purchase-tests-password"}, -1, 200)
		cookies[i] = response.Result().Cookies()[0]
		var envelope struct {
			Data struct {
				CSRF string `json:"csrf_token"`
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		csrf[i] = envelope.Data.CSRF
	}

	var proof bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: uint8(20 * x), G: uint8(20 * y), B: 90, A: 255})
		}
	}
	if err := png.Encode(&proof, img); err != nil {
		t.Fatal(err)
	}
	proofData := "data:image/png;base64," + base64.StdEncoding.EncodeToString(proof.Bytes())
	purchasePath := "/api/v1/me/purchases"
	optionsPath := fmt.Sprintf("/api/v1/me/packages/%d/payment-options", item.ID)
	request("GET", optionsPath, nil, -1, 401)
	request("GET", optionsPath, nil, 0, 403)
	request("GET", optionsPath, nil, 4, 403)
	request("GET", optionsPath, nil, 5, 403)
	request("GET", optionsPath, nil, 2, 200)
	request("POST", purchasePath, map[string]any{}, -1, 401)
	request("POST", purchasePath, map[string]any{}, 0, 403)
	request("POST", purchasePath, map[string]any{}, 4, 403)
	request("POST", purchasePath, map[string]any{}, 5, 403)
	request("POST", purchasePath, map[string]any{"departure_id": departures[0].ID, "national_adults": 1, "payment_method": "transfer"}, 2, 422)

	payload := map[string]any{
		"departure_id": departures[0].ID, "payment_method": "transfer", "national_adults": 1, "foreign_adults": 1,
		"minors": []map[string]any{{"age": 5, "is_foreign": false}, {"age": 8, "is_foreign": true}}, "payment_proof": proofData,
	}
	response := request("POST", purchasePath, payload, 2, 201)
	var created struct {
		Data models.TourPackagePurchase `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	purchase := created.Data
	if purchase.Status != models.PurchasePaymentReview || purchase.CapacityCount != 3 || purchase.FreeMinorCount != 1 || purchase.PayingMinorCount != 1 || purchase.TotalCents != 85000 || len(purchase.Minors) != 2 || purchase.Minors[0].Pays || !purchase.Minors[1].Pays {
		t.Fatalf("wrong purchase calculation: %+v", purchase)
	}
	if err := db.First(&departures[0], departures[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[0].HeldCapacity != 3 || departures[0].ConfirmedCapacity != 0 {
		t.Fatalf("payment review did not hold capacity: %+v", departures[0])
	}
	request("GET", fmt.Sprintf("/api/v1/me/purchases/%d", purchase.ID), nil, 3, 404)
	request("GET", fmt.Sprintf("/api/v1/agency/purchases/%d", purchase.ID), nil, 1, 404)
	request("GET", fmt.Sprintf("/api/v1/agency/purchases/%d/proof", purchase.ID), nil, 0, 200)
	reviewPath := fmt.Sprintf("/api/v1/agency/purchases/%d/review", purchase.ID)
	request("PATCH", reviewPath, map[string]any{"version": purchase.Version, "decision": "confirm"}, -1, 401)
	request("PATCH", reviewPath, map[string]any{"version": purchase.Version, "decision": "confirm"}, 2, 403)
	request("PATCH", reviewPath, map[string]any{"version": purchase.Version, "decision": "confirm"}, 4, 403)
	request("PATCH", reviewPath, map[string]any{"version": purchase.Version, "decision": "confirm"}, 5, 403)
	request("PATCH", reviewPath, map[string]any{"version": purchase.Version, "decision": "confirm"}, 1, 404)
	confirmedResponse := request("PATCH", reviewPath, map[string]any{"version": purchase.Version, "decision": "confirm"}, 0, 200)
	if err := json.Unmarshal(confirmedResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Status != models.PurchaseConfirmed {
		t.Fatalf("purchase not confirmed: %s", confirmedResponse.Body.String())
	}
	if err := db.First(&departures[0], departures[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[0].HeldCapacity != 0 || departures[0].ConfirmedCapacity != 3 || departures[0].Status != "open" {
		t.Fatalf("confirmed capacity is wrong before minimum: %+v", departures[0])
	}
	request("PATCH", reviewPath, map[string]any{"version": purchase.Version, "decision": "confirm"}, 0, 409)

	secondPayload := map[string]any{"departure_id": departures[0].ID, "payment_method": "transfer", "national_adults": 1, "foreign_adults": 0, "minors": []any{}, "payment_proof": proofData}
	secondResponse := request("POST", purchasePath, secondPayload, 3, 201)
	if err := json.Unmarshal(secondResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	request("PATCH", fmt.Sprintf("/api/v1/agency/purchases/%d/review", created.Data.ID), map[string]any{"version": created.Data.Version, "decision": "confirm"}, 0, 200)
	if err := db.First(&departures[0], departures[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[0].ConfirmedCapacity != 4 || departures[0].Status != "confirmed" {
		t.Fatalf("minimum was not reached correctly: %+v", departures[0])
	}
	request("POST", purchasePath, map[string]any{"departure_id": departures[0].ID, "payment_method": "transfer", "national_adults": 7, "foreign_adults": 0, "minors": []any{}, "payment_proof": proofData}, 2, 409)

	rejectPayload := map[string]any{"departure_id": departures[2].ID, "payment_method": "transfer", "national_adults": 2, "foreign_adults": 0, "minors": []any{}, "payment_proof": proofData}
	rejectResponse := request("POST", purchasePath, rejectPayload, 2, 201)
	if err := json.Unmarshal(rejectResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	correctionResponse := request("PATCH", fmt.Sprintf("/api/v1/agency/purchases/%d/review", created.Data.ID), map[string]any{"version": created.Data.Version, "decision": "request_correction", "reason": "Comprobante ilegible"}, 0, 200)
	if err := json.Unmarshal(correctionResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Status != models.PurchaseCorrectionRequested {
		t.Fatalf("correction was not requested: %s", correctionResponse.Body.String())
	}
	if err := db.First(&departures[2], departures[2].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[2].HeldCapacity != 2 {
		t.Fatalf("correction request released capacity: %+v", departures[2])
	}
	proofPath := fmt.Sprintf("/api/v1/me/purchases/%d/proof", created.Data.ID)
	request("PATCH", proofPath, map[string]any{"version": created.Data.Version, "payment_proof": proofData}, -1, 401)
	request("PATCH", proofPath, map[string]any{"version": created.Data.Version, "payment_proof": proofData}, 0, 403)
	request("PATCH", proofPath, map[string]any{"version": created.Data.Version, "payment_proof": proofData}, 4, 403)
	request("PATCH", proofPath, map[string]any{"version": created.Data.Version, "payment_proof": proofData}, 5, 403)
	request("PATCH", proofPath, map[string]any{"version": created.Data.Version, "payment_proof": proofData}, 3, 404)
	newProofResponse := request("PATCH", proofPath, map[string]any{"version": created.Data.Version, "payment_proof": proofData}, 2, 200)
	if err := json.Unmarshal(newProofResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Status != models.PurchasePaymentReview {
		t.Fatalf("corrected proof did not return to review: %s", newProofResponse.Body.String())
	}
	refundResponse := request("PATCH", fmt.Sprintf("/api/v1/agency/purchases/%d/review", created.Data.ID), map[string]any{"version": created.Data.Version, "decision": "refund", "reason": "El pago requiere devolución"}, 0, 200)
	if err := json.Unmarshal(refundResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Status != models.PurchaseRefundPending {
		t.Fatalf("refund was not registered: %s", refundResponse.Body.String())
	}
	if err := db.First(&departures[2], departures[2].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[2].HeldCapacity != 0 || departures[2].ConfirmedCapacity != 0 {
		t.Fatalf("refund did not release capacity: %+v", departures[2])
	}
	refundDestinationPath := fmt.Sprintf("/api/v1/me/purchases/%d/refund-destination", created.Data.ID)
	completeRefundWithoutDestinationPath := fmt.Sprintf("/api/v1/agency/purchases/%d/refund", created.Data.ID)
	request("PATCH", completeRefundWithoutDestinationPath, map[string]any{"version": created.Data.Version, "refund_proof": ""}, 0, 422)
	request("PATCH", completeRefundWithoutDestinationPath, map[string]any{"version": created.Data.Version, "refund_proof": proofData}, 0, 409)
	bankDestination := map[string]any{"version": created.Data.Version, "refund_method": "bank_transfer", "refund_bank_name": "Banco de prueba", "refund_account_holder": "Prueba Compra", "refund_account_number": "987654"}
	request("PATCH", refundDestinationPath, map[string]any{"version": created.Data.Version, "refund_method": "bank_transfer", "refund_bank_name": "", "refund_account_holder": "", "refund_account_number": ""}, 2, 422)
	request("PATCH", refundDestinationPath, bankDestination, -1, 401)
	request("PATCH", refundDestinationPath, bankDestination, 0, 403)
	request("PATCH", refundDestinationPath, bankDestination, 3, 404)
	bankResponse := request("PATCH", refundDestinationPath, bankDestination, 2, 200)
	if err := json.Unmarshal(bankResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.RefundMethod != "bank_transfer" || created.Data.RefundBankName != "Banco de prueba" || created.Data.RefundAccountNumber != "987654" {
		t.Fatalf("bank refund destination was not saved: %+v", created.Data)
	}

	cancelPayload := map[string]any{"departure_id": departures[3].ID, "payment_method": "transfer", "national_adults": 1, "foreign_adults": 0, "minors": []any{}, "payment_proof": proofData}
	cancelPurchaseResponse := request("POST", purchasePath, cancelPayload, 3, 201)
	if err := json.Unmarshal(cancelPurchaseResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	cancelPurchaseID := created.Data.ID
	if err := db.First(&departures[3], departures[3].ID).Error; err != nil {
		t.Fatal(err)
	}
	request("POST", fmt.Sprintf("/api/v1/agency/packages/%d/departures/%d/cancel", item.ID, departures[3].ID), map[string]any{"version": departures[3].Version, "reason": "Condiciones climáticas adversas"}, 0, 200)
	var cancelledPurchase models.TourPackagePurchase
	if err := db.First(&cancelledPurchase, cancelPurchaseID).Error; err != nil {
		t.Fatal(err)
	}
	if cancelledPurchase.Status != models.PurchaseRefundPending || cancelledPurchase.ReviewedAt == nil || cancelledPurchase.RefundReason == "" || cancelledPurchase.RefundDueAt == nil {
		t.Fatalf("cancelled departure did not create refund tracking: %+v", cancelledPurchase)
	}
	if err := db.First(&departures[3], departures[3].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[3].Status != "cancelled" || departures[3].HeldCapacity != 0 || departures[3].ConfirmedCapacity != 0 {
		t.Fatalf("cancelled departure retained capacity: %+v", departures[3])
	}

	touristCancelPayload := map[string]any{"departure_id": departures[5].ID, "payment_method": "transfer", "national_adults": 1, "foreign_adults": 0, "minors": []any{}, "payment_proof": proofData}
	touristCancelPurchase := request("POST", purchasePath, touristCancelPayload, 2, 201)
	if err := json.Unmarshal(touristCancelPurchase.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	touristCancelReview := fmt.Sprintf("/api/v1/agency/purchases/%d/review", created.Data.ID)
	confirmedForCancel := request("PATCH", touristCancelReview, map[string]any{"version": created.Data.Version, "decision": "confirm"}, 0, 200)
	if err := json.Unmarshal(confirmedForCancel.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	touristCancelPath := fmt.Sprintf("/api/v1/me/purchases/%d/cancel", created.Data.ID)
	touristCancelBody := map[string]any{"version": created.Data.Version, "reason": "Cambio de planes personales", "refund_method": "qr", "refund_qr": proofData}
	request("POST", touristCancelPath, touristCancelBody, -1, 401)
	request("POST", touristCancelPath, touristCancelBody, 0, 403)
	request("POST", touristCancelPath, touristCancelBody, 5, 403)
	request("POST", touristCancelPath, touristCancelBody, 3, 404)
	touristCancelled := request("POST", touristCancelPath, touristCancelBody, 2, 200)
	if err := json.Unmarshal(touristCancelled.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Status != models.PurchaseRefundPending || created.Data.RefundMethod != "qr" || created.Data.CancelledAt == nil || created.Data.RefundRequestedAt == nil || created.Data.RefundDueAt == nil || created.Data.RefundDueAt.Sub(*created.Data.RefundRequestedAt) != 72*time.Hour {
		t.Fatalf("tourist cancellation did not initialize refund: %+v", created.Data)
	}
	refundQRPath := fmt.Sprintf("/api/v1/me/purchases/%d/refund-qr", created.Data.ID)
	request("GET", refundQRPath, nil, 3, 404)
	request("GET", refundQRPath, nil, 2, 200)
	request("GET", fmt.Sprintf("/api/v1/agency/purchases/%d/refund-qr", created.Data.ID), nil, 1, 404)
	request("GET", fmt.Sprintf("/api/v1/agency/purchases/%d/refund-qr", created.Data.ID), nil, 0, 200)
	if err := db.First(&departures[5], departures[5].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[5].HeldCapacity != 0 || departures[5].ConfirmedCapacity != 0 {
		t.Fatalf("tourist cancellation did not release capacity: %+v", departures[5])
	}
	completeRefundPath := fmt.Sprintf("/api/v1/agency/purchases/%d/refund", created.Data.ID)
	completeRefundBody := map[string]any{"version": created.Data.Version, "refund_proof": proofData, "refund_reference": "DEV-REFUND-001"}
	request("PATCH", completeRefundPath, completeRefundBody, -1, 401)
	request("PATCH", completeRefundPath, completeRefundBody, 2, 403)
	request("PATCH", completeRefundPath, completeRefundBody, 4, 403)
	request("PATCH", completeRefundPath, completeRefundBody, 5, 403)
	request("PATCH", completeRefundPath, completeRefundBody, 1, 404)
	completedRefund := request("PATCH", completeRefundPath, completeRefundBody, 0, 200)
	if err := json.Unmarshal(completedRefund.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Status != models.PurchaseRefunded || created.Data.RefundedAt == nil || created.Data.RefundCompletedByUserID == nil || created.Data.RefundReference != "DEV-REFUND-001" {
		t.Fatalf("refund completion was not recorded: %+v", created.Data)
	}
	request("GET", fmt.Sprintf("/api/v1/me/purchases/%d/refund-proof", created.Data.ID), nil, 3, 404)
	request("GET", fmt.Sprintf("/api/v1/me/purchases/%d/refund-proof", created.Data.ID), nil, 2, 200)
	request("GET", fmt.Sprintf("/api/v1/agency/purchases/%d/refund-proof", created.Data.ID), nil, 1, 404)
	request("GET", fmt.Sprintf("/api/v1/agency/purchases/%d/refund-proof", created.Data.ID), nil, 0, 200)
	request("PATCH", completeRefundPath, completeRefundBody, 0, 409)

	minimumPayload := map[string]any{"departure_id": departures[4].ID, "payment_method": "transfer", "national_adults": 1, "foreign_adults": 0, "minors": []any{}, "payment_proof": proofData}
	minimumPurchaseResponse := request("POST", purchasePath, minimumPayload, 3, 201)
	if err := json.Unmarshal(minimumPurchaseResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	minimumPurchaseID := created.Data.ID
	request("PATCH", fmt.Sprintf("/api/v1/agency/purchases/%d/review", minimumPurchaseID), map[string]any{"version": created.Data.Version, "decision": "confirm"}, 0, 200)
	if err := db.Model(&models.TourPackageDeparture{}).Where("id = ?", departures[4].ID).Update("booking_closes_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	marked, err := services.MarkMinimumReviews(db, time.Now())
	if err != nil || marked < 1 {
		t.Fatalf("minimum review was not evaluated: count=%d err=%v", marked, err)
	}
	if err := db.First(&departures[4], departures[4].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[4].Status != "minimum_review" || departures[4].MinimumReviewAt == nil || departures[4].ConfirmedCapacity != 1 {
		t.Fatalf("departure did not wait for agency confirmation: %+v", departures[4])
	}
	minimumRefundPath := fmt.Sprintf("/api/v1/agency/packages/%d/departures/%d/minimum-refund", item.ID, departures[4].ID)
	request("POST", minimumRefundPath, map[string]any{"version": departures[4].Version}, -1, 401)
	request("POST", minimumRefundPath, map[string]any{"version": departures[4].Version}, 2, 403)
	request("POST", minimumRefundPath, map[string]any{"version": departures[4].Version}, 4, 403)
	request("POST", minimumRefundPath, map[string]any{"version": departures[4].Version}, 5, 403)
	request("POST", minimumRefundPath, map[string]any{"version": departures[4].Version}, 1, 404)
	request("POST", minimumRefundPath, map[string]any{"version": departures[4].Version}, 0, 200)
	var minimumPurchase models.TourPackagePurchase
	if err := db.First(&minimumPurchase, minimumPurchaseID).Error; err != nil {
		t.Fatal(err)
	}
	if minimumPurchase.Status != models.PurchaseRefundPending || minimumPurchase.RefundDueAt == nil {
		t.Fatalf("minimum failure did not start refund: %+v", minimumPurchase)
	}
	if err := db.First(&departures[4], departures[4].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[4].Status != "cancelled" || departures[4].MinimumReviewedAt == nil || departures[4].ConfirmedCapacity != 0 {
		t.Fatalf("minimum refund was not confirmed: %+v", departures[4])
	}

	latePayload := map[string]any{"departure_id": departures[6].ID, "payment_method": "transfer", "national_adults": 1, "foreign_adults": 0, "minors": []any{}, "payment_proof": proofData}
	latePurchaseResponse := request("POST", purchasePath, latePayload, 2, 201)
	if err := json.Unmarshal(latePurchaseResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	lateReviewPath := fmt.Sprintf("/api/v1/agency/purchases/%d/review", created.Data.ID)
	lateConfirmed := request("PATCH", lateReviewPath, map[string]any{"version": created.Data.Version, "decision": "confirm"}, 0, 200)
	if err := json.Unmarshal(lateConfirmed.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	lateCancelPath := fmt.Sprintf("/api/v1/me/purchases/%d/cancel", created.Data.ID)
	request("POST", lateCancelPath, map[string]any{"version": created.Data.Version, "reason": "Solicitud fuera del plazo permitido", "refund_method": "qr", "refund_qr": proofData}, 2, 409)
	if err := db.First(&departures[6], departures[6].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[6].ConfirmedCapacity != 1 {
		t.Fatalf("late cancellation changed capacity: %+v", departures[6])
	}

	lastSeat := map[string]any{"departure_id": departures[1].ID, "payment_method": "transfer", "national_adults": 1, "foreign_adults": 0, "minors": []any{}, "payment_proof": proofData}
	start := make(chan struct{})
	responses := make(chan *httptest.ResponseRecorder, 2)
	for _, account := range []int{2, 3} {
		go func(account int) {
			<-start
			responses <- rawRequest("POST", purchasePath, lastSeat, account)
		}(account)
	}
	close(start)
	statusCounts := map[int]int{}
	for range 2 {
		statusCounts[(<-responses).Code]++
	}
	if statusCounts[http.StatusCreated] != 1 || statusCounts[http.StatusConflict] != 1 {
		t.Fatalf("concurrent last-seat purchase got statuses: %v", statusCounts)
	}
	if err := db.First(&departures[1], departures[1].ID).Error; err != nil {
		t.Fatal(err)
	}
	if departures[1].HeldCapacity != 1 || departures[1].HeldCapacity+departures[1].ConfirmedCapacity > departures[1].MaxCapacity {
		t.Fatalf("last seat was oversold: %+v", departures[1])
	}
}
