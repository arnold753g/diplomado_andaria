package handlers

import (
	"testing"
	"time"
)

func TestPackagePublicationRequirementsAndItineraryReuse(t *testing.T) {
	draft := packageInput{Name: "Ruta del vino", DurationDays: 1}
	if fields := validatePackageInput(&draft, true); len(fields) != 0 {
		t.Fatalf("a minimal draft should be valid: %v", fields)
	}

	missing := packageInput{Name: "Ruta del vino", DurationDays: 1, Published: true}
	fields := validatePackageInput(&missing, true)
	if fields["published"] == "" {
		t.Fatal("an incomplete published package was accepted")
	}

	valid := packageInput{
		Name: "Ruta del vino", Description: "Una experiencia completa por los viñedos de Tarija.",
		DurationDays: 2, DurationNights: 1, NationalPriceCents: 15000, Published: true,
		Photos: []packagePhotoInput{{Data: "placeholder"}},
		Itinerary: []packageDayInput{
			{DayNumber: 2, Title: "Segundo día", AttractionIDs: []uint64{12}},
			{DayNumber: 1, Title: "Primer día", AttractionIDs: []uint64{12}},
		},
	}
	futureDate := time.Now().In(boliviaTime).AddDate(0, 0, 2).Format("2006-01-02")
	valid.Schedule = &packageScheduleInput{FrequencyType: packageFrequencyDaily, ValidFrom: futureDate, ValidUntil: futureDate, DepartureTime: "08:30", MeetingTime: "08:00", DefaultMinCapacity: 1, DefaultMaxCapacity: 10, BookingCutoffHours: 1, MaximumAdvanceDays: 30}
	if fields := validatePackageInput(&valid, true); len(fields) != 0 {
		t.Fatalf("the same attraction must be allowed on different days: %v", fields)
	}
	if valid.Itinerary[0].DayNumber != 1 || valid.Itinerary[1].DayNumber != 2 {
		t.Fatal("itinerary days were not sorted")
	}

	valid.Itinerary[0].AttractionIDs = []uint64{12, 12}
	if fields := validatePackageInput(&valid, true); fields["itinerary"] == "" {
		t.Fatal("a repeated attraction within one day was accepted")
	}
}

func TestPackageCancellationPolicyIsConsistent(t *testing.T) {
	input := packageInput{Name: "City tour", DurationDays: 1, CancellationNoticeHours: 48}
	if fields := validatePackageInput(&input, true); len(fields) != 0 || input.CancellationNoticeHours != 0 {
		t.Fatalf("disabled cancellation should clear its deadline: fields=%v hours=%d", fields, input.CancellationNoticeHours)
	}
	input.CancellationAllowed = true
	if fields := validatePackageInput(&input, true); fields["cancellation_notice_hours"] == "" {
		t.Fatal("enabled cancellation without notice was accepted")
	}
}
