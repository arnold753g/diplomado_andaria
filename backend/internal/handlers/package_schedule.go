package handlers

import (
	"errors"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/models"
)

const (
	packageFrequencySingle   = "single"
	packageFrequencyDaily    = "daily"
	packageFrequencyWeekdays = "specific_weekdays"
)

var boliviaTime = time.FixedZone("America/La_Paz", -4*60*60)

type packageScheduleInput struct {
	ID                  uint64 `json:"id"`
	FrequencyType       string `json:"frequency_type"`
	ValidFrom           string `json:"valid_from"`
	ValidUntil          string `json:"valid_until"`
	Weekdays            []int  `json:"weekdays"`
	DepartureTime       string `json:"departure_time"`
	MeetingTime         string `json:"meeting_time"`
	DefaultMinCapacity  int    `json:"default_min_capacity"`
	DefaultMaxCapacity  int    `json:"default_max_capacity"`
	BookingCutoffHours  int    `json:"booking_cutoff_hours"`
	MaximumAdvanceDays  int    `json:"maximum_advance_days"`
	DefaultMeetingPoint string `json:"default_meeting_point"`
	DefaultInstructions string `json:"default_instructions"`
}

func normalizeClock(value string) (string, int, int, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"15:04", "15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Format("15:04"), parsed.Hour(), parsed.Minute(), nil
		}
	}
	return "", 0, 0, errors.New("invalid time")
}

func parseScheduleDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

func scheduleToday(value time.Time) time.Time {
	local := value.In(boliviaTime)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func validatePackageSchedule(input *packageScheduleInput, fields map[string]string, create, published bool, now time.Time) {
	if input == nil {
		if published {
			fields["schedule"] = "Configura la frecuencia y las fechas antes de publicar"
		}
		return
	}
	input.FrequencyType = strings.TrimSpace(input.FrequencyType)
	if !slices.Contains([]string{packageFrequencySingle, packageFrequencyDaily, packageFrequencyWeekdays}, input.FrequencyType) {
		fields["schedule.frequency_type"] = "Selecciona una frecuencia válida"
	}
	from, fromErr := parseScheduleDate(input.ValidFrom)
	until, untilErr := parseScheduleDate(input.ValidUntil)
	if fromErr != nil || untilErr != nil {
		fields["schedule.dates"] = "Selecciona fechas válidas"
	} else {
		if input.FrequencyType == packageFrequencySingle {
			until = from
			input.ValidUntil = input.ValidFrom
		}
		if until.Before(from) || until.After(from.AddDate(1, 0, 0)) {
			fields["schedule.dates"] = "El rango debe finalizar después de su inicio y no superar doce meses"
		}
		if create && from.Before(scheduleToday(now)) {
			fields["schedule.dates"] = "Una nueva programación no puede comenzar en una fecha pasada"
		}
	}
	departure, departureHour, departureMinute, departureErr := normalizeClock(input.DepartureTime)
	if departureErr != nil {
		fields["schedule.departure_time"] = "Selecciona una hora de salida válida"
	} else {
		input.DepartureTime = departure
	}
	if strings.TrimSpace(input.MeetingTime) == "" {
		input.MeetingTime = input.DepartureTime
	}
	meeting, meetingHour, meetingMinute, meetingErr := normalizeClock(input.MeetingTime)
	if meetingErr != nil {
		fields["schedule.meeting_time"] = "Selecciona una hora de encuentro válida"
	} else {
		input.MeetingTime = meeting
	}
	if departureErr == nil && meetingErr == nil && (meetingHour > departureHour || (meetingHour == departureHour && meetingMinute > departureMinute)) {
		fields["schedule.meeting_time"] = "La hora de encuentro no puede ser posterior a la salida"
	}
	if input.DefaultMinCapacity < 1 || input.DefaultMinCapacity > 500 || input.DefaultMaxCapacity < input.DefaultMinCapacity || input.DefaultMaxCapacity > 500 {
		fields["schedule.capacity"] = "Define cupos válidos entre 1 y 500"
	}
	if input.BookingCutoffHours < 1 || input.BookingCutoffHours > 720 {
		fields["schedule.booking_cutoff_hours"] = "El cierre de compra debe estar entre 1 y 720 horas"
	}
	if input.MaximumAdvanceDays < 1 || input.MaximumAdvanceDays > 365 || input.BookingCutoffHours >= input.MaximumAdvanceDays*24 {
		fields["schedule.maximum_advance_days"] = "La anticipación máxima debe estar entre 1 y 365 días y superar el cierre de compra"
	}
	input.DefaultMeetingPoint = strings.TrimSpace(input.DefaultMeetingPoint)
	input.DefaultInstructions = strings.TrimSpace(input.DefaultInstructions)
	if len([]rune(input.DefaultMeetingPoint)) > 500 {
		fields["schedule.default_meeting_point"] = "El punto de encuentro admite hasta 500 caracteres"
	}
	if len([]rune(input.DefaultInstructions)) > 3000 {
		fields["schedule.default_instructions"] = "Las instrucciones admiten hasta 3000 caracteres"
	}
	if input.FrequencyType == packageFrequencyWeekdays {
		seen := map[int]bool{}
		days := make([]int, 0, len(input.Weekdays))
		for _, day := range input.Weekdays {
			if day < 1 || day > 7 || seen[day] {
				fields["schedule.weekdays"] = "Selecciona días de semana válidos sin repetir"
				continue
			}
			seen[day] = true
			days = append(days, day)
		}
		slices.Sort(days)
		input.Weekdays = days
		if len(days) == 0 {
			fields["schedule.weekdays"] = "Selecciona al menos un día de la semana"
		}
	} else {
		input.Weekdays = []int{}
	}
	if published && input.FrequencyType == packageFrequencySingle && fromErr == nil && departureErr == nil {
		startsAt := time.Date(from.Year(), from.Month(), from.Day(), departureHour, departureMinute, 0, 0, boliviaTime)
		if !startsAt.After(now) {
			fields["schedule.dates"] = "La salida única debe tener una fecha y hora futuras"
		} else if !startsAt.Add(-time.Duration(input.BookingCutoffHours) * time.Hour).After(now) {
			fields["schedule.booking_cutoff_hours"] = "La salida única ya se encuentra fuera de su plazo de compra"
		}
	}
}

func scheduleMatchesDay(schedule *models.TourPackageSchedule, day time.Time) bool {
	switch schedule.FrequencyType {
	case packageFrequencySingle:
		return day.Equal(schedule.ValidFrom)
	case packageFrequencyDaily:
		return true
	case packageFrequencyWeekdays:
		isoDay := int(day.Weekday())
		if isoDay == 0 {
			isoDay = 7
		}
		return slices.Contains(schedule.Weekdays, isoDay)
	default:
		return false
	}
}

func buildPackageDepartures(schedule *models.TourPackageSchedule, now time.Time) ([]models.TourPackageDeparture, error) {
	_, departureHour, departureMinute, err := normalizeClock(schedule.DepartureTime)
	if err != nil {
		return nil, err
	}
	_, meetingHour, meetingMinute, err := normalizeClock(schedule.MeetingTime)
	if err != nil {
		return nil, err
	}
	result := make([]models.TourPackageDeparture, 0)
	for day := schedule.ValidFrom; !day.After(schedule.ValidUntil); day = day.AddDate(0, 0, 1) {
		if !scheduleMatchesDay(schedule, day) {
			continue
		}
		startsAt := time.Date(day.Year(), day.Month(), day.Day(), departureHour, departureMinute, 0, 0, boliviaTime)
		if !startsAt.After(now) {
			continue
		}
		meetingAt := time.Date(day.Year(), day.Month(), day.Day(), meetingHour, meetingMinute, 0, 0, boliviaTime)
		bookingOpensAt := startsAt.AddDate(0, 0, -schedule.MaximumAdvanceDays)
		bookingClosesAt := startsAt.Add(-time.Duration(schedule.BookingCutoffHours) * time.Hour)
		status := "open"
		if !bookingClosesAt.After(now) {
			status = "closed"
		}
		scheduleID := schedule.ID
		result = append(result, models.TourPackageDeparture{
			PackageID: schedule.PackageID, ScheduleID: &scheduleID, StartsAt: startsAt, MeetingAt: meetingAt,
			BookingOpensAt: bookingOpensAt, BookingClosesAt: bookingClosesAt,
			MinCapacity: schedule.DefaultMinCapacity, MaxCapacity: schedule.DefaultMaxCapacity,
			MeetingPoint: schedule.DefaultMeetingPoint, Instructions: schedule.DefaultInstructions,
			Status: status, Version: 1,
		})
	}
	return result, nil
}

func savePackageSchedule(tx *gorm.DB, packageID uint64, input *packageScheduleInput, now time.Time) error {
	if input == nil {
		if err := tx.Where("package_id = ? AND starts_at > ? AND held_capacity = 0 AND confirmed_capacity = 0 AND is_exception = FALSE", packageID, now).Delete(&models.TourPackageDeparture{}).Error; err != nil {
			return err
		}
		return tx.Where("package_id = ?", packageID).Delete(&models.TourPackageSchedule{}).Error
	}
	from, _ := parseScheduleDate(input.ValidFrom)
	until, _ := parseScheduleDate(input.ValidUntil)
	var schedule models.TourPackageSchedule
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("package_id = ?", packageID).First(&schedule).Error
	creating := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !creating {
		return err
	}
	if creating {
		schedule.PackageID = packageID
	}
	schedule.FrequencyType = input.FrequencyType
	schedule.ValidFrom = from
	schedule.ValidUntil = until
	schedule.Weekdays = input.Weekdays
	schedule.DepartureTime = input.DepartureTime
	schedule.MeetingTime = input.MeetingTime
	schedule.DefaultMinCapacity = input.DefaultMinCapacity
	schedule.DefaultMaxCapacity = input.DefaultMaxCapacity
	schedule.BookingCutoffHours = input.BookingCutoffHours
	schedule.MaximumAdvanceDays = input.MaximumAdvanceDays
	schedule.DefaultMeetingPoint = input.DefaultMeetingPoint
	schedule.DefaultInstructions = input.DefaultInstructions
	if creating {
		if err := tx.Create(&schedule).Error; err != nil {
			return err
		}
	} else if err := tx.Save(&schedule).Error; err != nil {
		return err
	}
	if err := tx.Where("package_id = ? AND starts_at > ? AND held_capacity = 0 AND confirmed_capacity = 0 AND is_exception = FALSE", packageID, now).Delete(&models.TourPackageDeparture{}).Error; err != nil {
		return err
	}
	departures, err := buildPackageDepartures(&schedule, now)
	if err != nil {
		return err
	}
	for i := range departures {
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "package_id"}, {Name: "starts_at"}}, DoNothing: true}).Create(&departures[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
