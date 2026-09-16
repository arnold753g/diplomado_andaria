package handlers

import (
	"testing"
	"time"

	"starter-backend/internal/models"
)

func validDepartureForUpdate() (models.TourPackageDeparture, time.Time) {
	now := time.Date(2026, 9, 21, 10, 0, 0, 0, boliviaTime)
	start := time.Date(2026, 9, 28, 8, 30, 0, 0, boliviaTime)
	return models.TourPackageDeparture{
		StartsAt: start, BookingOpensAt: start.AddDate(0, 0, -30),
		HeldCapacity: 2, ConfirmedCapacity: 3, Version: 1,
	}, now
}

func TestValidateDepartureUpdateAcceptsAvailableFutureWindow(t *testing.T) {
	departure, now := validDepartureForUpdate()
	input := departureUpdateInput{
		Version: 1, MeetingTime: "08:00", MinCapacity: 4, MaxCapacity: 12,
		BookingCutoffHours: 24, MeetingPoint: "  Plaza principal  ", Instructions: " Llegar temprano ",
	}
	meetingAt, bookingClosesAt, fields := validateDepartureUpdate(&input, departure, now)
	if len(fields) != 0 {
		t.Fatalf("valid adjustment rejected: %v", fields)
	}
	if meetingAt.Hour() != 8 || meetingAt.Minute() != 0 {
		t.Fatalf("wrong meeting time: %s", meetingAt)
	}
	if !bookingClosesAt.Equal(departure.StartsAt.Add(-24 * time.Hour)) {
		t.Fatalf("wrong purchase cutoff: %s", bookingClosesAt)
	}
	if input.MeetingPoint != "Plaza principal" || input.Instructions != "Llegar temprano" {
		t.Fatalf("text was not normalized: %+v", input)
	}
}

func TestValidateDepartureUpdateProtectsOccupiedCapacity(t *testing.T) {
	departure, now := validDepartureForUpdate()
	input := departureUpdateInput{Version: 1, MeetingTime: "08:00", MinCapacity: 1, MaxCapacity: 4, BookingCutoffHours: 24}
	_, _, fields := validateDepartureUpdate(&input, departure, now)
	if fields["max_capacity"] == "" {
		t.Fatalf("occupied capacity reduction accepted: %v", fields)
	}
}

func TestValidateDepartureUpdateRejectsInvalidTimes(t *testing.T) {
	departure, now := validDepartureForUpdate()
	input := departureUpdateInput{Version: 1, MeetingTime: "09:00", MinCapacity: 1, MaxCapacity: 10, BookingCutoffHours: 24}
	_, _, fields := validateDepartureUpdate(&input, departure, now)
	if fields["meeting_time"] == "" {
		t.Fatalf("meeting after departure accepted: %v", fields)
	}

	input.MeetingTime = "08:00"
	input.BookingCutoffHours = 200
	_, _, fields = validateDepartureUpdate(&input, departure, now)
	if fields["booking_cutoff_hours"] == "" {
		t.Fatalf("elapsed purchase window accepted: %v", fields)
	}
}
