package httpapi

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
	"testing"
	"time"
)

// Explicitly gated: synthetic data in a temporary PostgreSQL schema for browser QA.
func TestAgencyPaymentsBrowserFixture(t *testing.T) {
	if os.Getenv("PAYMENTS_UI_FIXTURE") != "true" {
		t.Skip("browser fixture disabled")
	}
	password := os.Getenv("PAYMENTS_UI_PASSWORD")
	if err := security.ValidatePassword(password); err != nil {
		t.Fatal("set a synthetic PAYMENTS_UI_PASSWORD for browser QA")
	}
	db := testutil.Database(t)
	cfg := testConfig()
	cfg.APIRatePerMinute = 10000
	cfg.AllowedOrigins = []string{"http://127.0.0.1:3319"}
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
		users[i].PasswordHash, _ = security.HashPassword(password, 10)
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
			d.AuthorizedAt = &now
			d.AuthorizedByUserID = &users[0].ID
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

	for i := 0; i < 28; i++ {
		p := models.TourPackagePurchase{Reference: fmt.Sprintf("PAY-%03d", i), TouristID: users[2].ID, AgencyID: agencies[0].ID, PackageID: packages[0].ID, DepartureID: departures[0].ID, Status: models.PurchaseConfirmed, PaymentMethod: "qr", NationalAdults: 1, CapacityCount: 1, NationalUnitPriceCents: 10000, TotalCents: 10000, TouristName: fmt.Sprintf("Comprador %03d", i), TouristEmail: "comprador@example.test", TouristPhone: "000000", PaymentProof: []byte{1, 2}, ReviewedAt: &now, Version: 1, Minors: []models.PurchaseMinor{}}
		if i == 0 {
			p.Minors = []models.PurchaseMinor{{Age: 3}}
			p.FreeMinorCount = 1
			p.TouristName = "Comprador de prueba 01"
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

	}

	fixture := http.NewServeMux()
	fixture.HandleFunc("GET /__payments_fixture", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("andaria-payments-fixture-v1")) })
	fixture.Handle("/", handler)
	server := &http.Server{Addr: "127.0.0.1:18886", Handler: fixture, ReadHeaderTimeout: 5 * time.Second}
	t.Cleanup(func() { server.Close() })
	errors := make(chan error, 1)
	go func() { errors <- server.ListenAndServe() }()
	fmt.Println("PAYMENTS_FIXTURE_READY http://127.0.0.1:18886")
	stop := os.Getenv("PAYMENTS_UI_STOP_FILE")
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-errors:
			t.Fatal(err)
		case <-ticker.C:
			if _, err := os.Stat(stop); err == nil {
				return
			}
		}
	}
}
