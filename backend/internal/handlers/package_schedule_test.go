package handlers

import (
	"testing"
	"time"

	"starter-backend/internal/models"
)

func TestSpecificWeekdayScheduleBuildsConcreteDepartures(t *testing.T) {
	now := time.Date(2026, 9, 21, 10, 0, 0, 0, boliviaTime)
	input := packageScheduleInput{
		FrequencyType: packageFrequencyWeekdays, ValidFrom: "2026-09-21", ValidUntil: "2026-09-27",
		Weekdays: []int{4, 2}, DepartureTime: "08:30", MeetingTime: "08:00",
		DefaultMinCapacity: 4, DefaultMaxCapacity: 12, BookingCutoffHours: 24, MaximumAdvanceDays: 180,
	}
	fields := map[string]string{}
	validatePackageSchedule(&input, fields, true, true, now)
	if len(fields) != 0 {
		t.Fatalf("valid schedule rejected: %v", fields)
	}
	if len(input.Weekdays) != 2 || input.Weekdays[0] != 2 || input.Weekdays[1] != 4 {
		t.Fatalf("weekdays were not normalized: %v", input.Weekdays)
	}
	from, _ := parseScheduleDate(input.ValidFrom)
	until, _ := parseScheduleDate(input.ValidUntil)
	schedule := models.TourPackageSchedule{
		ID: 3, PackageID: 9, FrequencyType: input.FrequencyType, ValidFrom: from, ValidUntil: until,
		Weekdays: input.Weekdays, DepartureTime: input.DepartureTime, MeetingTime: input.MeetingTime,
		DefaultMinCapacity: 4, DefaultMaxCapacity: 12, BookingCutoffHours: 24, MaximumAdvanceDays: 180,
	}
	departures, err := buildPackageDepartures(&schedule, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(departures) != 2 {
		t.Fatalf("got %d departures, want Tuesday and Thursday", len(departures))
	}
	if departures[0].StartsAt.Weekday() != time.Tuesday || departures[1].StartsAt.Weekday() != time.Thursday {
		t.Fatalf("wrong generated weekdays: %s, %s", departures[0].StartsAt.Weekday(), departures[1].StartsAt.Weekday())
	}
	if departures[0].BookingClosesAt != departures[0].StartsAt.Add(-24*time.Hour) {
		t.Fatal("booking cutoff was not copied to the departure")
	}
	if departures[0].BookingOpensAt != departures[0].StartsAt.AddDate(0, 0, -180) {
		t.Fatal("maximum booking advance was not copied to the departure")
	}
}

func TestSingleDepartureMustBeFutureWhenPublished(t *testing.T) {
	now := time.Date(2026, 9, 21, 10, 0, 0, 0, boliviaTime)
	input := packageScheduleInput{
		FrequencyType: packageFrequencySingle, ValidFrom: "2026-09-21", ValidUntil: "2026-09-21",
		DepartureTime: "09:00", MeetingTime: "08:30", DefaultMinCapacity: 1, DefaultMaxCapacity: 8,
		BookingCutoffHours: 1, MaximumAdvanceDays: 30,
	}
	fields := map[string]string{}
	validatePackageSchedule(&input, fields, true, true, now)
	if fields["schedule.dates"] == "" {
		t.Fatal("published package accepted a past single departure")
	}
	input.DepartureTime = "10:30"
	input.MeetingTime = "10:00"
	fields = map[string]string{}
	validatePackageSchedule(&input, fields, true, true, now)
	if fields["schedule.booking_cutoff_hours"] == "" {
		t.Fatal("published package accepted a single departure whose purchase window was already closed")
	}
}

func TestScheduleRangeAndPurchaseWindowAreBounded(t *testing.T) {
	input := packageScheduleInput{
		FrequencyType: packageFrequencyDaily, ValidFrom: "2027-01-01", ValidUntil: "2028-01-02",
		DepartureTime: "08:00", MeetingTime: "07:30", DefaultMinCapacity: 1, DefaultMaxCapacity: 8,
		BookingCutoffHours: 48, MaximumAdvanceDays: 2,
	}
	fields := map[string]string{}
	validatePackageSchedule(&input, fields, true, false, time.Date(2026, 9, 21, 10, 0, 0, 0, boliviaTime))
	if fields["schedule.dates"] == "" || fields["schedule.maximum_advance_days"] == "" {
		t.Fatalf("unbounded schedule was accepted: %v", fields)
	}
}
