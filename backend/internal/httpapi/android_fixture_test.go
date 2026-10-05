package httpapi

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
)

// Opt-in fixture for Android integration tests. testutil owns an isolated schema
// and drops it after the stop file is created or the bounded lifetime expires.
func TestAndroidFixture(t *testing.T) {
	if os.Getenv("ANDARIA_ANDROID_FIXTURE") != "1" {
		t.Skip("manual Android fixture; set ANDARIA_ANDROID_FIXTURE=1")
	}
	stopFile := os.Getenv("ANDARIA_ANDROID_STOP_FILE")
	if stopFile == "" {
		t.Fatal("ANDARIA_ANDROID_STOP_FILE is required")
	}
	if _, err := os.Stat(stopFile); err == nil {
		t.Fatal("stop file already exists; choose a fresh filename")
	}
	db := testutil.Database(t)
	password, err := security.HashPassword("andaria-android-test", 10)
	if err != nil {
		t.Fatal(err)
	}
	accounts := []models.User{
		{Email: "android-tourist@example.test", Role: models.RoleUser},
		{Email: "android-agency@example.test", Role: models.RoleAgency},
		{Email: "android-attraction@example.test", Role: models.RoleAttraction},
	}
	for i := range accounts {
		accounts[i].FirstName, accounts[i].LastName = "Android", "Prueba"
		accounts[i].Status, accounts[i].PasswordHash = models.StatusActive, password
		accounts[i].Phone, accounts[i].DocumentNumber, accounts[i].Nationality = "70000000", "PRUEBA-ANDROID", "Boliviana"
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	qr, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGNYtfsMAAREAjLFENlZAAAAAElFTkSuQmCC")
	if err != nil {
		t.Fatal(err)
	}
	agency := models.Agency{Name: "Agencia Android de prueba", Department: "Tarija", City: "Tarija", Address: "Centro", Phone: "70000000", Email: "android-agency@example.test", ManagerID: accounts[1].ID, Status: models.StatusActive, Published: true, MinimumPayingAge: 6, AcceptsQR: true, QRImage: qr, AcceptsTransfer: true, BankName: "Banco de prueba", AccountHolder: "Agencia de prueba", AccountNumber: "PRUEBA-123", Version: 1}
	if err := db.Create(&agency).Error; err != nil {
		t.Fatal(err)
	}
	item := models.TourPackage{AgencyID: agency.ID, Name: "Experiencia Android de prueba", Description: "Experiencia aislada para verificar la app Android.", DurationDays: 1, NationalPriceCents: 25000, ForeignSurchargeCents: 5000, CancellationAllowed: true, CancellationNoticeHours: 48, Published: true, Version: 1, Includes: []string{"Guía de prueba"}, Excludes: []string{}, Bring: []string{"Agua"}}
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	departure := models.TourPackageDeparture{PackageID: item.ID, StartsAt: now.Add(14 * 24 * time.Hour), MeetingAt: now.Add(14*24*time.Hour - time.Hour), BookingOpensAt: now.Add(-time.Hour), BookingClosesAt: now.Add(13 * 24 * time.Hour), MinCapacity: 1, MaxCapacity: 20, MeetingPoint: "Plaza de prueba", Status: "open", Version: 1}
	if err := db.Create(&departure).Error; err != nil {
		t.Fatal(err)
	}
	attraction := models.Attraction{Name: "Mirador Android de prueba", Description: "Lugar de prueba para favoritos.", Department: "Tarija", City: "Tarija", Address: "Centro", ManagerID: accounts[2].ID, Status: models.StatusActive, Published: true, ScheduleMode: "all_day", OpeningDays: []int{1, 2, 3, 4, 5, 6, 7}, SeasonMode: "all_year", Version: 1}
	if err := db.Create(&attraction).Error; err != nil {
		t.Fatal(err)
	}
	cfg := testConfig()
	cfg.AppEnv, cfg.RegistrationEnabled = "test", true
	cfg.AuthRatePerMinute, cfg.APIRatePerMinute = 1000, 10000
	handler, err := New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		t.Fatal(err)
	}
	fixtureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/__android_fixture" {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = io.WriteString(w, "andaria-android-fixture-v1")
			return
		}
		handler.ServeHTTP(w, r)
	})
	server := &http.Server{Handler: fixtureHandler, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			t.Error(err)
		}
	})
	t.Log("Android fixture ready on 127.0.0.1:8081; only isolated test data")
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(25 * time.Minute)
	defer deadline.Stop()
	for {
		select {
		case <-deadline.C:
			return
		case <-ticker.C:
			if _, err := os.Stat(stopFile); err == nil {
				return
			}
		}
	}
}
