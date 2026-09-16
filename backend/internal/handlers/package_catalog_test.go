package handlers

import (
	"testing"
	"time"

	"starter-backend/internal/models"
)

func TestPublicDepartureSeparatesAvailabilityFromMinimum(t *testing.T) {
	now := time.Date(2026, 9, 22, 9, 0, 0, 0, boliviaTime)
	departure := models.TourPackageDeparture{
		ID: 7, StartsAt: now.AddDate(0, 0, 5), MeetingAt: now.AddDate(0, 0, 5).Add(-30 * time.Minute),
		BookingOpensAt: now.Add(-time.Hour), BookingClosesAt: now.AddDate(0, 0, 4),
		MinCapacity: 6, MaxCapacity: 12, HeldCapacity: 2, ConfirmedCapacity: 3, Status: "open",
	}
	view := publicDepartureView(departure, now)
	if !view.Bookable || view.AvailableCapacity != 7 || view.MinimumReached || view.RemainingForMinimum != 3 {
		t.Fatalf("wrong public availability: %+v", view)
	}
}

func TestPublicDepartureClosesOutsidePurchaseWindow(t *testing.T) {
	now := time.Date(2026, 9, 22, 9, 0, 0, 0, boliviaTime)
	departure := models.TourPackageDeparture{
		StartsAt: now.AddDate(0, 0, 5), BookingOpensAt: now.Add(time.Hour), BookingClosesAt: now.AddDate(0, 0, 4),
		MinCapacity: 1, MaxCapacity: 3, Status: "open",
	}
	if publicDepartureView(departure, now).Bookable {
		t.Fatal("departure became bookable before its purchase window")
	}
	departure.BookingOpensAt = now.Add(-time.Hour)
	departure.HeldCapacity = 2
	departure.ConfirmedCapacity = 1
	if publicDepartureView(departure, now).Bookable {
		t.Fatal("sold-out departure remained bookable")
	}
}
