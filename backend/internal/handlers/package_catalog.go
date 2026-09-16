package handlers

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
)

type publicPackageDeparture struct {
	ID                  uint64    `json:"id"`
	StartsAt            time.Time `json:"starts_at"`
	MeetingAt           time.Time `json:"meeting_at"`
	BookingOpensAt      time.Time `json:"booking_opens_at"`
	BookingClosesAt     time.Time `json:"booking_closes_at"`
	MinCapacity         int       `json:"min_capacity"`
	MaxCapacity         int       `json:"max_capacity"`
	ConfirmedCapacity   int       `json:"confirmed_capacity"`
	AvailableCapacity   int       `json:"available_capacity"`
	RemainingForMinimum int       `json:"remaining_for_minimum"`
	MinimumReached      bool      `json:"minimum_reached"`
	MeetingPoint        string    `json:"meeting_point"`
	Instructions        string    `json:"instructions"`
	Status              string    `json:"status"`
	Bookable            bool      `json:"bookable"`
}

type publicPackageSummary struct {
	ID                    uint64                    `json:"id"`
	AgencyID              uint64                    `json:"agency_id"`
	AgencyName            string                    `json:"agency_name"`
	AgencyDepartment      string                    `json:"agency_department"`
	AgencyCity            string                    `json:"agency_city"`
	Name                  string                    `json:"name"`
	Description           string                    `json:"description"`
	DurationDays          int                       `json:"duration_days"`
	DurationNights        int                       `json:"duration_nights"`
	Difficulty            string                    `json:"difficulty"`
	NationalPriceCents    int                       `json:"national_price_cents"`
	ForeignSurchargeCents int                       `json:"foreign_surcharge_cents"`
	Photos                []models.TourPackagePhoto `json:"photos"`
	FrequencyType         string                    `json:"frequency_type"`
	NextDeparture         *publicPackageDeparture   `json:"next_departure,omitempty"`
	Bookable              bool                      `json:"bookable"`
}

type publicPackageDetail struct {
	publicPackageSummary
	Includes                []string                     `json:"includes"`
	Excludes                []string                     `json:"excludes"`
	Bring                   []string                     `json:"bring"`
	CancellationAllowed     bool                         `json:"cancellation_allowed"`
	CancellationNoticeHours int                          `json:"cancellation_notice_hours"`
	MinimumPayingAge        int                          `json:"minimum_paying_age"`
	Itinerary               []models.PackageItineraryDay `json:"itinerary"`
	Departures              []publicPackageDeparture     `json:"departures"`
}

func packagePublicScope(db *gorm.DB) *gorm.DB {
	return db.Model(&models.TourPackage{}).
		Joins("JOIN agencies catalog_agency ON catalog_agency.id = tour_packages.agency_id").
		Where("tour_packages.published = TRUE AND catalog_agency.status = ? AND catalog_agency.published = TRUE", models.StatusActive)
}

func packageCatalogQuery(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Agency", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "name", "department", "city", "minimum_paying_age")
		}).
		Preload("Photos", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "package_id", "position").Order("position, id")
		}).
		Preload("Schedule", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "package_id", "frequency_type")
		})
}

func packagePublicDetailQuery(db *gorm.DB, now time.Time) *gorm.DB {
	return packageCatalogQuery(db).
		Preload("Itinerary", func(tx *gorm.DB) *gorm.DB { return tx.Order("day_number, id") }).
		Preload("Itinerary.Attractions", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("EXISTS (SELECT 1 FROM attractions a WHERE a.id = tour_package_itinerary_attractions.attraction_id AND a.status = ? AND a.published = TRUE)", models.StatusActive).Order("position")
		}).
		Preload("Itinerary.Attractions.Attraction", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "name", "department", "city")
		}).
		Preload("Departures", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("starts_at > ? AND status IN ?", now, []string{"open", "confirmed", "closed"}).Order("starts_at, id").Limit(400)
		})
}

func publicDepartureView(departure models.TourPackageDeparture, now time.Time) publicPackageDeparture {
	available := departure.MaxCapacity - departure.HeldCapacity - departure.ConfirmedCapacity
	if available < 0 {
		available = 0
	}
	remainingForMinimum := departure.MinCapacity - departure.ConfirmedCapacity
	if remainingForMinimum < 0 {
		remainingForMinimum = 0
	}
	bookableStatus := departure.Status == "open" || departure.Status == "confirmed"
	bookable := bookableStatus && !now.Before(departure.BookingOpensAt) && now.Before(departure.BookingClosesAt) && departure.StartsAt.After(now) && available > 0
	return publicPackageDeparture{
		ID: departure.ID, StartsAt: departure.StartsAt, MeetingAt: departure.MeetingAt,
		BookingOpensAt: departure.BookingOpensAt, BookingClosesAt: departure.BookingClosesAt,
		MinCapacity: departure.MinCapacity, MaxCapacity: departure.MaxCapacity,
		ConfirmedCapacity: departure.ConfirmedCapacity, AvailableCapacity: available,
		RemainingForMinimum: remainingForMinimum, MinimumReached: remainingForMinimum == 0,
		MeetingPoint: departure.MeetingPoint, Instructions: departure.Instructions,
		Status: departure.Status, Bookable: bookable,
	}
}

func publicPackageSummaryView(item models.TourPackage, departures []models.TourPackageDeparture, now time.Time) publicPackageSummary {
	view := publicPackageSummary{
		ID: item.ID, AgencyID: item.AgencyID, Name: item.Name, Description: item.Description,
		DurationDays: item.DurationDays, DurationNights: item.DurationNights, Difficulty: item.Difficulty,
		NationalPriceCents: item.NationalPriceCents, ForeignSurchargeCents: item.ForeignSurchargeCents,
		Photos: item.Photos,
	}
	if item.Agency != nil {
		view.AgencyName = item.Agency.Name
		view.AgencyDepartment = item.Agency.Department
		view.AgencyCity = item.Agency.City
	}
	if item.Schedule != nil {
		view.FrequencyType = item.Schedule.FrequencyType
	}
	for _, departure := range departures {
		candidate := publicDepartureView(departure, now)
		if view.NextDeparture == nil {
			copy := candidate
			view.NextDeparture = &copy
		}
		if candidate.Bookable {
			if !view.Bookable {
				copy := candidate
				view.NextDeparture = &copy
			}
			view.Bookable = true
		}
	}
	return view
}

func catalogDate(value string, end bool) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, err
	}
	result := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, boliviaTime)
	if end {
		result = result.AddDate(0, 0, 1)
	}
	return result, nil
}

func catalogIntFilter(r *http.Request, name string, min, max int, fields map[string]string) *int {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < min || parsed > max {
		fields[name] = "Selecciona un valor válido"
		return nil
	}
	return &parsed
}

func (h *PackageHandler) PublicOptions(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]any{
		"departments": departments,
		"difficulties": []map[string]string{
			{"label": "Fácil", "value": "easy"},
			{"label": "Moderada", "value": "moderate"},
			{"label": "Exigente", "value": "demanding"},
		},
	}, "")
}

func (h *PackageHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	page, limit := queryInt(r, "page", 1, 1, 100000), queryInt(r, "limit", 12, 1, 50)
	fields := map[string]string{}
	minDuration := catalogIntFilter(r, "min_duration", 1, 30, fields)
	maxDuration := catalogIntFilter(r, "max_duration", 1, 30, fields)
	maxPrice := catalogIntFilter(r, "max_price_cents", 1, 100000000, fields)
	if minDuration != nil && maxDuration != nil && *maxDuration < *minDuration {
		fields["duration"] = "La duración máxima debe ser igual o mayor que la mínima"
	}
	difficulty := strings.TrimSpace(r.URL.Query().Get("difficulty"))
	if difficulty != "" && !slices.Contains([]string{"easy", "moderate", "demanding"}, difficulty) {
		fields["difficulty"] = "Selecciona una dificultad válida"
	}
	department := strings.TrimSpace(r.URL.Query().Get("department"))
	if department != "" && !slices.Contains(departments, department) {
		fields["department"] = "Selecciona un departamento válido"
	}
	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sort == "" {
		sort = "next_departure"
	}
	if !slices.Contains([]string{"next_departure", "price_asc", "price_desc", "newest"}, sort) {
		fields["sort"] = "Selecciona un orden válido"
	}
	var dateFrom, dateUntil *time.Time
	if value := strings.TrimSpace(r.URL.Query().Get("date_from")); value != "" {
		parsed, err := catalogDate(value, false)
		if err != nil {
			fields["date_from"] = "Selecciona una fecha inicial válida"
		} else {
			dateFrom = &parsed
		}
	}
	if value := strings.TrimSpace(r.URL.Query().Get("date_until")); value != "" {
		parsed, err := catalogDate(value, true)
		if err != nil {
			fields["date_until"] = "Selecciona una fecha final válida"
		} else {
			dateUntil = &parsed
		}
	}
	if dateFrom != nil && dateUntil != nil && !dateUntil.After(*dateFrom) {
		fields["dates"] = "La fecha final debe ser igual o posterior a la inicial"
	}
	bookableOnly := strings.TrimSpace(r.URL.Query().Get("bookable"))
	if bookableOnly != "" && bookableOnly != "true" && bookableOnly != "false" {
		fields["bookable"] = "El filtro de disponibilidad no es válido"
	}
	if len(fields) > 0 {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PACKAGE_FILTER_INVALID", "Revisa los filtros del catálogo", fields)
		return
	}

	q := packagePublicScope(h.db)
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		pattern := "%" + escapeLike(strings.ToLower(search)) + "%"
		q = q.Where("LOWER(tour_packages.name || ' ' || tour_packages.description || ' ' || catalog_agency.name || ' ' || catalog_agency.city) LIKE ? ESCAPE '!'", pattern)
	}
	if department != "" {
		q = q.Where("catalog_agency.department = ?", department)
	}
	if difficulty != "" {
		q = q.Where("tour_packages.difficulty = ?", difficulty)
	}
	if minDuration != nil {
		q = q.Where("tour_packages.duration_days >= ?", *minDuration)
	}
	if maxDuration != nil {
		q = q.Where("tour_packages.duration_days <= ?", *maxDuration)
	}
	if maxPrice != nil {
		q = q.Where("tour_packages.national_price_cents <= ?", *maxPrice)
	}
	departureConditions := "d.package_id = tour_packages.id AND d.starts_at > ? AND d.status IN ('open','confirmed','closed')"
	departureArgs := []any{now}
	if dateFrom != nil {
		departureConditions += " AND d.starts_at >= ?"
		departureArgs = append(departureArgs, *dateFrom)
	}
	if dateUntil != nil {
		departureConditions += " AND d.starts_at < ?"
		departureArgs = append(departureArgs, *dateUntil)
	}
	if dateFrom != nil || dateUntil != nil {
		q = q.Where("EXISTS (SELECT 1 FROM tour_package_departures d WHERE "+departureConditions+")", departureArgs...)
	}
	if bookableOnly == "true" {
		q = q.Where("EXISTS (SELECT 1 FROM tour_package_departures d WHERE d.package_id = tour_packages.id AND d.starts_at > ? AND d.booking_opens_at <= ? AND d.booking_closes_at > ? AND d.status IN ('open','confirmed') AND d.held_capacity + d.confirmed_capacity < d.max_capacity)", now, now, now)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		packageError(w, r, err)
		return
	}
	switch sort {
	case "price_asc":
		q = q.Order("tour_packages.national_price_cents ASC, tour_packages.id DESC")
	case "price_desc":
		q = q.Order("tour_packages.national_price_cents DESC, tour_packages.id DESC")
	case "newest":
		q = q.Order("tour_packages.created_at DESC, tour_packages.id DESC")
	default:
		q = q.Order("(SELECT MIN(d.starts_at) FROM tour_package_departures d WHERE d.package_id = tour_packages.id AND d.starts_at > NOW() AND d.status IN ('open','confirmed','closed')) ASC NULLS LAST, tour_packages.id DESC")
	}
	var packages []models.TourPackage
	if err := packageCatalogQuery(q).Limit(limit).Offset((page - 1) * limit).Find(&packages).Error; err != nil {
		packageError(w, r, err)
		return
	}

	ids := make([]uint64, 0, len(packages))
	for _, item := range packages {
		ids = append(ids, item.ID)
	}
	departuresByPackage := map[uint64][]models.TourPackageDeparture{}
	if len(ids) > 0 {
		var departures []models.TourPackageDeparture
		if err := h.db.Where("package_id IN ? AND starts_at > ? AND status IN ?", ids, now, []string{"open", "confirmed", "closed"}).Order("package_id, starts_at, id").Find(&departures).Error; err != nil {
			packageError(w, r, err)
			return
		}
		for _, departure := range departures {
			departuresByPackage[departure.PackageID] = append(departuresByPackage[departure.PackageID], departure)
		}
	}
	views := make([]publicPackageSummary, 0, len(packages))
	for _, item := range packages {
		views = append(views, publicPackageSummaryView(item, departuresByPackage[item.ID], now))
	}
	respond.JSON(w, http.StatusOK, map[string]any{
		"packages":   views,
		"pagination": map[string]any{"page": page, "limit": limit, "total": total},
	}, "")
}

func (h *PackageHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	now := time.Now()
	var item models.TourPackage
	if err := packagePublicDetailQuery(packagePublicScope(h.db), now).First(&item, id).Error; err != nil {
		packageError(w, r, err)
		return
	}
	departures := make([]publicPackageDeparture, 0, len(item.Departures))
	for _, departure := range item.Departures {
		departures = append(departures, publicDepartureView(departure, now))
	}
	summary := publicPackageSummaryView(item, item.Departures, now)
	minimumPayingAge := 0
	if item.Agency != nil {
		minimumPayingAge = item.Agency.MinimumPayingAge
	}
	respond.JSON(w, http.StatusOK, publicPackageDetail{
		publicPackageSummary: summary,
		Includes:             item.Includes, Excludes: item.Excludes, Bring: item.Bring,
		CancellationAllowed: item.CancellationAllowed, CancellationNoticeHours: item.CancellationNoticeHours,
		MinimumPayingAge: minimumPayingAge, Itinerary: item.Itinerary, Departures: departures,
	}, "")
}

func (h *PackageHandler) PhotoPublic(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	photoID, err := strconv.ParseUint(mux.Vars(r)["photo"], 10, 64)
	if err != nil || photoID == 0 {
		packageError(w, r, gorm.ErrRecordNotFound)
		return
	}
	var photo models.TourPackagePhoto
	permitted := packagePublicScope(h.db).Select("tour_packages.id").Where("tour_packages.id = ?", id)
	if err := h.db.Where("id = ? AND package_id IN (?)", photoID, permitted).First(&photo).Error; err != nil {
		packageError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(photo.Image)
}
