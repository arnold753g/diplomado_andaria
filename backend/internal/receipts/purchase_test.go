package receipts

import (
	"bytes"
	"os"
	"path/filepath"
	"starter-backend/internal/models"
	"strings"
	"testing"
	"time"
)

func TestReceiptPricesAndUnicodeLayout(t *testing.T) {
	now := time.Date(2026, 10, 28, 16, 0, 0, 0, time.UTC)
	p := models.TourPackagePurchase{
		ID: 1, Reference: "AND-DEMO-0001", Status: models.PurchaseConfirmed, PaymentMethod: "transfer",
		NationalAdults: 1, ForeignAdults: 1, Minors: []models.PurchaseMinor{{Age: 3}, {Age: 10, IsForeign: true, Pays: true}},
		FreeMinorCount: 1, PayingMinorCount: 1, CapacityCount: 3, NationalUnitPriceCents: 25000, ForeignSurchargeCents: 5000, TotalCents: 85000,
		TouristName: "María Muñoz — Turista de prueba", TouristDocument: "DOC-DEMO", TouristEmail: "turista@example.test", TouristPhone: "00000000", CreatedAt: now, PaymentConfirmedAt: &now,
		ReceiptSnapshot: models.PurchaseReceiptSnapshot{Origin: "compra", CapturedAt: now, AgencyName: "Agencia de prueba Andaria", AgencyAddress: "Dirección ficticia, Tarija", AgencyPhone: "00000000", AgencyEmail: "agencia@example.test", PackageName: "Aventura en Tariquía", DurationDays: 2, DurationNights: 1, DepartureAt: now.Add(10 * 24 * time.Hour), MeetingAt: now.Add(10*24*time.Hour - 15*time.Minute), MeetingPoint: "Punto de encuentro de prueba", CancellationAllowed: true, CancellationNoticeHours: 24, Bring: []string{"Documento de identidad", "Agua y protección solar"}},
	}
	sum, count := 0, 0
	for _, r := range rows(p) {
		sum += r.count * r.unit
		count += r.count
	}
	if sum != 85000 || count != 4 {
		t.Fatalf("wrong original tariffs: total=%d tourists=%d", sum, count)
	}
	departure := models.TourPackageDeparture{StartsAt: p.ReceiptSnapshot.DepartureAt, Status: "confirmed", MinCapacity: 3, ConfirmedCapacity: 3}
	for _, state := range []string{models.PurchaseConfirmed, models.PurchaseRefundPending, models.PurchaseRefunded} {
		p.Status = state
		if state != models.PurchaseConfirmed {
			departure.Status = "cancelled"
			p.RefundReason = "Cancelación de prueba por condiciones climáticas"
			due := now.Add(72 * time.Hour)
			p.RefundDueAt = &due
		}
		if state == models.PurchaseRefunded {
			p.RefundedAt = &now
			p.RefundReference = "DEV-DEMO-01"
		}
		data, err := Purchase(p, departure, true, now)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasPrefix(data, []byte("%PDF-")) || !bytes.Contains(data, []byte("%%EOF")) {
			t.Fatal("invalid PDF output")
		}
		if output := os.Getenv("RECEIPT_SAMPLE_DIR"); output != "" {
			if err = os.MkdirAll(output, 0700); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(output, state+".pdf"), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	p.ReceiptSnapshot.PackageName = strings.Repeat("Paquete extenso con tildes y ñ ", 40)
	p.ReceiptSnapshot.MeetingPoint = strings.Repeat("p", 500)
	p.ReceiptSnapshot.Bring = []string{strings.Repeat("Recomendaciones de prueba para comprobar saltos de página. ", 100)}
	data, err := Purchase(p, departure, false, now)
	if err != nil {
		t.Fatal(err)
	}
	if output := os.Getenv("RECEIPT_SAMPLE_DIR"); output != "" {
		if err = os.WriteFile(filepath.Join(output, "texto-largo.pdf"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, state := range []string{models.PurchasePaymentReview, models.PurchaseCorrectionRequested, models.PurchaseCancelled, models.PurchasePaymentRejected} {
		if Available(state) {
			t.Fatalf("receipt available before financial decision: %s", state)
		}
	}
}
