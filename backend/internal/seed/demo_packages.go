package seed

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"time"

	"gorm.io/gorm"
	"starter-backend/internal/models"
)

type demoPackageSpec struct {
	Name, Description, Frequency           string
	DurationDays, DurationNights           int
	NationalPrice, ForeignSurcharge        int
	StartOffset, EndOffset                 int
	Weekdays                               []int
	DepartureHour, DepartureMinute         int
	MinCapacity, MaxCapacity               int
	BookingCutoffHours, MaximumAdvanceDays int
	MeetingPoint, Instructions             string
	ColorA, ColorB                         color.RGBA
}

func DemoPackages(db *gorm.DB, appEnv string) ([]models.TourPackage, error) {
	if appEnv == "production" {
		return nil, errors.New("demo packages are disabled in production")
	}
	var agency models.Agency
	if err := db.Where("status = ? AND published = TRUE", models.StatusActive).Order("id").First(&agency).Error; err != nil {
		return nil, fmt.Errorf("an active published agency is required: %w", err)
	}
	if !agency.AcceptsQR && !agency.AcceptsTransfer {
		agency.AcceptsTransfer = true
		agency.BankName = "Banco de prueba Andaria"
		agency.AccountHolder = agency.Name
		agency.AccountNumber = "000-ANDARIA-DEMO"
		agency.PaymentInstructions = "Realiza el pago completo y adjunta una fotografía legible del comprobante."
		agency.Version++
		if err := db.Save(&agency).Error; err != nil {
			return nil, err
		}
	}
	specs := []demoPackageSpec{
		{Name: "Ruta diaria de vinos y sabores", Description: "Recorrido diario por bodegas y sabores del valle central de Tarija.", Frequency: "daily", DurationDays: 1, NationalPrice: 32000, ForeignSurcharge: 6000, StartOffset: 5, EndOffset: 10, DepartureHour: 9, MinCapacity: 3, MaxCapacity: 8, BookingCutoffHours: 12, MaximumAdvanceDays: 30, MeetingPoint: "Plaza Luis de Fuentes", Instructions: "Presentarse 30 minutos antes con documento de identidad.", ColorA: color.RGBA{R: 91, G: 28, B: 54, A: 255}, ColorB: color.RGBA{R: 213, G: 152, B: 80, A: 255}},
		{Name: "Aventura de fin de semana en Tariquía", Description: "Experiencia de dos días en naturaleza con salidas los sábados y domingos.", Frequency: "specific_weekdays", DurationDays: 2, DurationNights: 1, NationalPrice: 68000, ForeignSurcharge: 12000, StartOffset: 6, EndOffset: 34, Weekdays: []int{0, 6}, DepartureHour: 7, DepartureMinute: 30, MinCapacity: 5, MaxCapacity: 14, BookingCutoffHours: 48, MaximumAdvanceDays: 90, MeetingPoint: "Terminal de buses de Tarija", Instructions: "Llevar abrigo, agua y calzado de caminata.", ColorA: color.RGBA{R: 18, G: 73, B: 52, A: 255}, ColorB: color.RGBA{R: 137, G: 190, B: 88, A: 255}},
		{Name: "Escapada única a San Lorenzo", Description: "Salida especial de cupos reducidos para conocer San Lorenzo y su gastronomía.", Frequency: "single", DurationDays: 1, NationalPrice: 24000, ForeignSurcharge: 4000, StartOffset: 14, EndOffset: 14, DepartureHour: 8, MinCapacity: 2, MaxCapacity: 5, BookingCutoffHours: 24, MaximumAdvanceDays: 60, MeetingPoint: "Parque Temático de Tarija", Instructions: "La movilidad sale puntualmente a las 08:00.", ColorA: color.RGBA{R: 45, G: 74, B: 97, A: 255}, ColorB: color.RGBA{R: 204, G: 171, B: 94, A: 255}},
	}
	created := make([]models.TourPackage, 0, len(specs))
	for _, spec := range specs {
		item, err := saveDemoPackage(db, agency, spec)
		if err != nil {
			return nil, err
		}
		created = append(created, item)
	}
	return created, nil
}

func saveDemoPackage(db *gorm.DB, agency models.Agency, spec demoPackageSpec) (models.TourPackage, error) {
	var result models.TourPackage
	err := db.Transaction(func(tx *gorm.DB) error {
		var existing models.TourPackage
		lookup := tx.Where("agency_id = ? AND name = ?", agency.ID, spec.Name).Limit(1).Find(&existing)
		if lookup.Error != nil {
			return lookup.Error
		}
		if lookup.RowsAffected > 0 {
			var purchases int64
			if err := tx.Model(&models.TourPackagePurchase{}).Where("package_id = ?", existing.ID).Count(&purchases).Error; err != nil {
				return err
			}
			if purchases > 0 {
				result = existing
				return nil
			}
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
		}
		result = models.TourPackage{
			AgencyID: agency.ID, Name: spec.Name, Description: spec.Description, DurationDays: spec.DurationDays, DurationNights: spec.DurationNights,
			Difficulty: "easy", NationalPriceCents: spec.NationalPrice, ForeignSurchargeCents: spec.ForeignSurcharge,
			Includes: []string{"Transporte desde el punto de encuentro", "Guía local"}, Excludes: []string{"Gastos personales"}, Bring: []string{"Documento de identidad", "Agua"},
			CancellationAllowed: true, CancellationNoticeHours: 72, Published: true, Version: 1,
		}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		photo, err := demoPhoto(spec.ColorA, spec.ColorB)
		if err != nil {
			return err
		}
		if err := tx.Create(&models.TourPackagePhoto{PackageID: result.ID, Position: 0, Image: photo}).Error; err != nil {
			return err
		}
		for day := 1; day <= spec.DurationDays; day++ {
			title := "Experiencia principal"
			if spec.DurationDays > 1 {
				title = fmt.Sprintf("Día %d de la aventura", day)
			}
			itinerary := models.PackageItineraryDay{PackageID: result.ID, DayNumber: day, Title: title, Description: spec.Description, Activities: []string{"Encuentro con el guía", "Recorrido acompañado", "Tiempo para fotografías"}}
			if err := tx.Create(&itinerary).Error; err != nil {
				return err
			}
		}
		zone := time.FixedZone("America/La_Paz", -4*60*60)
		now := time.Now().In(zone)
		startDate, endDate := now.AddDate(0, 0, spec.StartOffset), now.AddDate(0, 0, spec.EndOffset)
		weekdays := append([]int{}, spec.Weekdays...)
		schedule := models.TourPackageSchedule{
			PackageID: result.ID, FrequencyType: spec.Frequency,
			ValidFrom: time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, zone), ValidUntil: time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, zone),
			Weekdays: weekdays, DepartureTime: fmt.Sprintf("%02d:%02d:00", spec.DepartureHour, spec.DepartureMinute), MeetingTime: fmt.Sprintf("%02d:%02d:00", spec.DepartureHour, max(0, spec.DepartureMinute-30)),
			DefaultMinCapacity: spec.MinCapacity, DefaultMaxCapacity: spec.MaxCapacity, BookingCutoffHours: spec.BookingCutoffHours, MaximumAdvanceDays: spec.MaximumAdvanceDays,
			DefaultMeetingPoint: spec.MeetingPoint, DefaultInstructions: spec.Instructions,
		}
		if spec.DepartureMinute < 30 {
			meeting := time.Date(2000, 1, 1, spec.DepartureHour, spec.DepartureMinute, 0, 0, zone).Add(-30 * time.Minute)
			schedule.MeetingTime = meeting.Format("15:04:05")
		}
		if err := tx.Create(&schedule).Error; err != nil {
			return err
		}
		for date := schedule.ValidFrom; !date.After(schedule.ValidUntil); date = date.AddDate(0, 0, 1) {
			weekday := int(date.Weekday())
			if spec.Frequency == "specific_weekdays" && !containsWeekday(spec.Weekdays, weekday) {
				continue
			}
			startsAt := time.Date(date.Year(), date.Month(), date.Day(), spec.DepartureHour, spec.DepartureMinute, 0, 0, zone)
			departure := models.TourPackageDeparture{
				PackageID: result.ID, ScheduleID: &schedule.ID, StartsAt: startsAt, MeetingAt: startsAt.Add(-30 * time.Minute),
				BookingOpensAt: startsAt.AddDate(0, 0, -spec.MaximumAdvanceDays), BookingClosesAt: startsAt.Add(-time.Duration(spec.BookingCutoffHours) * time.Hour),
				MinCapacity: spec.MinCapacity, MaxCapacity: spec.MaxCapacity, MeetingPoint: spec.MeetingPoint, Instructions: spec.Instructions, Status: "open", Version: 1,
			}
			if err := tx.Create(&departure).Error; err != nil {
				return err
			}
			if spec.Frequency == "single" {
				break
			}
		}
		return nil
	})
	return result, err
}

func containsWeekday(values []int, value int) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func demoPhoto(a, b color.RGBA) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 1200, 720))
	for y := 0; y < 720; y++ {
		for x := 0; x < 1200; x++ {
			ratio := float64(x+y) / float64(1200+720)
			blend := func(start, end uint8) uint8 { return uint8(float64(start)*(1-ratio) + float64(end)*ratio) }
			img.SetRGBA(x, y, color.RGBA{R: blend(a.R, b.R), G: blend(a.G, b.G), B: blend(a.B, b.B), A: 255})
		}
	}
	var output bytes.Buffer
	if err := jpeg.Encode(&output, img, &jpeg.Options{Quality: 88}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func DemoPackageNames() []string {
	return []string{"Ruta diaria de vinos y sabores", "Aventura de fin de semana en Tariquía", "Escapada única a San Lorenzo"}
}

func IsDemoPackage(name string) bool {
	for _, candidate := range DemoPackageNames() {
		if strings.EqualFold(strings.TrimSpace(name), candidate) {
			return true
		}
	}
	return false
}
