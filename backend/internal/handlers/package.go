package handlers

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
)

var (
	errPackageConflict       = errors.New("package version conflict")
	errPackageAgencyInactive = errors.New("package agency inactive")
	errPackagePhoto          = errors.New("invalid package photo")
	errPackageAttraction     = errors.New("invalid package attraction")
)

type PackageHandler struct{ db *gorm.DB }

func NewPackageHandler(db *gorm.DB) *PackageHandler { return &PackageHandler{db: db} }

type packagePhotoInput struct {
	ID   uint64 `json:"id"`
	Data string `json:"data"`
}

type packageDayInput struct {
	DayNumber     int      `json:"day_number"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Activities    []string `json:"activities"`
	AttractionIDs []uint64 `json:"attraction_ids"`
}

type packageInput struct {
	Name                    string                `json:"name"`
	Description             string                `json:"description"`
	DurationDays            int                   `json:"duration_days"`
	DurationNights          int                   `json:"duration_nights"`
	Difficulty              string                `json:"difficulty"`
	NationalPriceCents      int                   `json:"national_price_cents"`
	ForeignSurchargeCents   int                   `json:"foreign_surcharge_cents"`
	Includes                []string              `json:"includes"`
	Excludes                []string              `json:"excludes"`
	Bring                   []string              `json:"bring"`
	CancellationAllowed     bool                  `json:"cancellation_allowed"`
	CancellationNoticeHours int                   `json:"cancellation_notice_hours"`
	Published               bool                  `json:"published"`
	Version                 int                   `json:"version"`
	Photos                  []packagePhotoInput   `json:"photos"`
	Itinerary               []packageDayInput     `json:"itinerary"`
	Schedule                *packageScheduleInput `json:"schedule"`
}

type packageResponse struct {
	models.TourPackage
	AgencyName       string `json:"agency_name"`
	MinimumPayingAge int    `json:"minimum_paying_age"`
}

func packageQuery(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Agency", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "name", "minimum_paying_age") }).
		Preload("Photos", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "package_id", "position").Order("position, id") }).
		Preload("Schedule").
		Preload("Itinerary", func(tx *gorm.DB) *gorm.DB { return tx.Order("day_number, id") }).
		Preload("Itinerary.Attractions", func(tx *gorm.DB) *gorm.DB { return tx.Order("position") }).
		Preload("Itinerary.Attractions.Attraction", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "name", "department", "city") })
}

func packageDetailQuery(db *gorm.DB) *gorm.DB {
	return packageQuery(db).Preload("Departures", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("starts_at >= ?", time.Now().Add(-24*time.Hour)).Order("starts_at, id").Limit(400)
	})
}

func packageView(p models.TourPackage) packageResponse {
	v := packageResponse{TourPackage: p}
	if p.Agency != nil {
		v.AgencyName = p.Agency.Name
		v.MinimumPayingAge = p.Agency.MinimumPayingAge
	}
	return v
}

func (h *PackageHandler) ownAgency(r *http.Request, db *gorm.DB) (models.Agency, error) {
	principal, _ := middleware.CurrentPrincipal(r)
	var agency models.Agency
	err := db.Where("manager_id = ?", principal.User.ID).First(&agency).Error
	return agency, err
}

func (h *PackageHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	agency, err := h.ownAgency(r, h.db)
	if err != nil {
		packageError(w, r, err)
		return
	}
	page, limit := queryInt(r, "page", 1, 1, 100000), queryInt(r, "limit", 12, 1, 50)
	q := h.db.Model(&models.TourPackage{}).Where("agency_id = ?", agency.ID)
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		q = q.Where("LOWER(name) LIKE ? ESCAPE '!'", "%"+escapeLike(strings.ToLower(search))+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		packageError(w, r, err)
		return
	}
	var packages []models.TourPackage
	if err := packageQuery(q).Order("tour_packages.created_at DESC, tour_packages.id DESC").Limit(limit).Offset((page - 1) * limit).Find(&packages).Error; err != nil {
		packageError(w, r, err)
		return
	}
	views := make([]packageResponse, 0, len(packages))
	for _, item := range packages {
		views = append(views, packageView(item))
	}
	respond.JSON(w, 200, map[string]any{"packages": views, "pagination": map[string]any{"page": page, "limit": limit, "total": total}}, "")
}

func (h *PackageHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	agency, err := h.ownAgency(r, h.db)
	if err != nil {
		packageError(w, r, err)
		return
	}
	var item models.TourPackage
	if err := packageDetailQuery(h.db.Where("agency_id = ?", agency.ID)).First(&item, id).Error; err != nil {
		packageError(w, r, err)
		return
	}
	respond.JSON(w, 200, packageView(item), "")
}

func (h *PackageHandler) Attractions(w http.ResponseWriter, r *http.Request) {
	if _, err := h.ownAgency(r, h.db); err != nil {
		packageError(w, r, err)
		return
	}
	var attractions []struct {
		ID         uint64 `json:"id"`
		Name       string `json:"name"`
		Department string `json:"department"`
		City       string `json:"city"`
	}
	q := h.db.Model(&models.Attraction{}).Select("id, name, department, city").Where("status = ? AND published = TRUE", models.StatusActive)
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		q = q.Where("LOWER(name || ' ' || city || ' ' || department) LIKE ? ESCAPE '!'", "%"+escapeLike(strings.ToLower(search))+"%")
	}
	if err := q.Order("name, id").Limit(100).Scan(&attractions).Error; err != nil {
		packageError(w, r, err)
		return
	}
	respond.JSON(w, 200, attractions, "")
}

func (h *PackageHandler) Create(w http.ResponseWriter, r *http.Request) { h.save(w, r, true) }
func (h *PackageHandler) Update(w http.ResponseWriter, r *http.Request) { h.save(w, r, false) }

func (h *PackageHandler) save(w http.ResponseWriter, r *http.Request, create bool) {
	var input packageInput
	if !decodeJSONLimit(w, r, &input, 42<<20) {
		return
	}
	fields := validatePackageInput(&input, create)
	images := map[int][]byte{}
	seenPhotos := map[uint64]bool{}
	if len(input.Photos) <= 6 {
		for i, photo := range input.Photos {
			if photo.ID > 0 {
				if create || photo.Data != "" || seenPhotos[photo.ID] {
					fields["photos"] = "Las fotografías existentes no son válidas"
				}
				seenPhotos[photo.ID] = true
				continue
			}
			image, err := decodeAttractionImage(photo.Data)
			if err != nil {
				fields["photos"] = err.Error()
				break
			}
			images[i] = image
		}
	}
	if len(fields) > 0 {
		respond.Error(w, r, 422, "PACKAGE_VALIDATION", "Revisa los datos del paquete", fields)
		return
	}

	var id uint64
	if !create {
		var ok bool
		id, ok = pathID(w, r)
		if !ok {
			return
		}
	}
	var item models.TourPackage
	err := h.db.Transaction(func(tx *gorm.DB) error {
		agency, err := h.ownAgency(r, tx.Clauses(clause.Locking{Strength: "UPDATE"}))
		if err != nil {
			return err
		}
		if agency.Status != models.StatusActive {
			return errPackageAgencyInactive
		}
		if create {
			item.AgencyID = agency.ID
			item.Version = 1
		} else {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("agency_id = ?", agency.ID).First(&item, id).Error; err != nil {
				return err
			}
			if item.Version != input.Version {
				return errPackageConflict
			}
			item.Version++
		}

		attractionIDs := make([]uint64, 0)
		for _, day := range input.Itinerary {
			attractionIDs = append(attractionIDs, day.AttractionIDs...)
		}
		if len(attractionIDs) > 0 {
			var count int64
			if err := tx.Model(&models.Attraction{}).Where("id IN ? AND status = ? AND published = TRUE", attractionIDs, models.StatusActive).Distinct("id").Count(&count).Error; err != nil {
				return err
			}
			unique := map[uint64]bool{}
			for _, attractionID := range attractionIDs {
				unique[attractionID] = true
			}
			if count != int64(len(unique)) {
				return errPackageAttraction
			}
		}

		item.Name = input.Name
		item.Description = input.Description
		item.DurationDays = input.DurationDays
		item.DurationNights = input.DurationNights
		item.Difficulty = input.Difficulty
		item.NationalPriceCents = input.NationalPriceCents
		item.ForeignSurchargeCents = input.ForeignSurchargeCents
		item.Includes = input.Includes
		item.Excludes = input.Excludes
		item.Bring = input.Bring
		item.CancellationAllowed = input.CancellationAllowed
		item.CancellationNoticeHours = input.CancellationNoticeHours
		item.Published = input.Published
		if create {
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		} else if err := tx.Save(&item).Error; err != nil {
			return err
		}

		keep := make([]uint64, 0, len(input.Photos))
		for position, photoInput := range input.Photos {
			if photoInput.ID > 0 {
				result := tx.Model(&models.TourPackagePhoto{}).Where("id = ? AND package_id = ?", photoInput.ID, item.ID).Update("position", position)
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected != 1 {
					return errPackagePhoto
				}
				keep = append(keep, photoInput.ID)
			} else {
				photo := models.TourPackagePhoto{PackageID: item.ID, Position: position, Image: images[position]}
				if err := tx.Create(&photo).Error; err != nil {
					return err
				}
				keep = append(keep, photo.ID)
			}
		}
		photoDelete := tx.Where("package_id = ?", item.ID)
		if len(keep) > 0 {
			photoDelete = photoDelete.Where("id NOT IN ?", keep)
		}
		if err := photoDelete.Delete(&models.TourPackagePhoto{}).Error; err != nil {
			return err
		}

		if err := tx.Where("package_id = ?", item.ID).Delete(&models.PackageItineraryDay{}).Error; err != nil {
			return err
		}
		for _, dayInput := range input.Itinerary {
			day := models.PackageItineraryDay{PackageID: item.ID, DayNumber: dayInput.DayNumber, Title: dayInput.Title, Description: dayInput.Description, Activities: dayInput.Activities}
			if err := tx.Create(&day).Error; err != nil {
				return err
			}
			for position, attractionID := range dayInput.AttractionIDs {
				link := models.PackageItineraryAttraction{ItineraryDayID: day.ID, AttractionID: attractionID, Position: position}
				if err := tx.Create(&link).Error; err != nil {
					return err
				}
			}
		}
		if err := savePackageSchedule(tx, item.ID, input.Schedule, time.Now()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		packageError(w, r, err)
		return
	}
	if err := packageDetailQuery(h.db).First(&item, item.ID).Error; err != nil {
		packageError(w, r, err)
		return
	}
	status := http.StatusOK
	if create {
		status = http.StatusCreated
	}
	respond.JSON(w, status, packageView(item), "Paquete guardado")
}

func (h *PackageHandler) PhotoMine(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	photoID, err := strconv.ParseUint(mux.Vars(r)["photo"], 10, 64)
	if err != nil {
		packageError(w, r, gorm.ErrRecordNotFound)
		return
	}
	agency, err := h.ownAgency(r, h.db)
	if err != nil {
		packageError(w, r, err)
		return
	}
	var photo models.TourPackagePhoto
	permitted := h.db.Model(&models.TourPackage{}).Select("id").Where("id = ? AND agency_id = ?", id, agency.ID)
	if err := h.db.Where("id = ? AND package_id IN (?)", photoID, permitted).First(&photo).Error; err != nil {
		packageError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(200)
	_, _ = w.Write(photo.Image)
}

func validatePackageInput(input *packageInput, create bool) map[string]string {
	fields := map[string]string{}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Difficulty = strings.TrimSpace(input.Difficulty)
	if input.Name == "" || len([]rune(input.Name)) > 160 {
		fields["name"] = "Escribe un nombre de hasta 160 caracteres"
	}
	if len([]rune(input.Description)) > 6000 {
		fields["description"] = "La descripción admite hasta 6000 caracteres"
	}
	if input.DurationDays < 1 || input.DurationDays > 30 {
		fields["duration_days"] = "La duración debe ser de 1 a 30 días"
	}
	if input.DurationNights < 0 || input.DurationNights >= input.DurationDays {
		fields["duration_nights"] = "Las noches deben ser menores que los días"
	}
	if !slices.Contains([]string{"", "easy", "moderate", "demanding"}, input.Difficulty) {
		fields["difficulty"] = "Selecciona una dificultad válida"
	}
	if input.NationalPriceCents < 0 || input.NationalPriceCents > 100000000 {
		fields["national_price_cents"] = "Ingresa un precio válido"
	}
	if input.ForeignSurchargeCents < 0 || input.ForeignSurchargeCents > 100000000 {
		fields["foreign_surcharge_cents"] = "Ingresa un recargo válido"
	}
	if input.CancellationAllowed {
		if input.CancellationNoticeHours < 1 || input.CancellationNoticeHours > 8760 {
			fields["cancellation_notice_hours"] = "Indica entre 1 y 8760 horas"
		}
	} else {
		input.CancellationNoticeHours = 0
	}
	if !create && input.Version < 1 {
		fields["version"] = "Recarga el paquete antes de guardar"
	}
	if len(input.Photos) > 6 {
		fields["photos"] = "Puedes guardar hasta seis fotografías"
	}
	input.Includes = normalizePackageList(input.Includes, "includes", fields)
	input.Excludes = normalizePackageList(input.Excludes, "excludes", fields)
	input.Bring = normalizePackageList(input.Bring, "bring", fields)
	seenDays := map[int]bool{}
	for i := range input.Itinerary {
		day := &input.Itinerary[i]
		day.Title = strings.TrimSpace(day.Title)
		day.Description = strings.TrimSpace(day.Description)
		if day.DayNumber < 1 || day.DayNumber > input.DurationDays || seenDays[day.DayNumber] {
			fields["itinerary"] = "Los días del itinerario deben ser únicos y estar dentro de la duración"
		}
		seenDays[day.DayNumber] = true
		if day.Title == "" || len([]rune(day.Title)) > 160 || len([]rune(day.Description)) > 3000 {
			fields["itinerary"] = "Completa títulos y descripciones válidos para el itinerario"
		}
		day.Activities = normalizePackageList(day.Activities, "itinerary", fields)
		seenAttractions := map[uint64]bool{}
		if len(day.AttractionIDs) > 20 {
			fields["itinerary"] = "Cada día admite hasta 20 atracciones"
		}
		for _, attractionID := range day.AttractionIDs {
			if attractionID == 0 || seenAttractions[attractionID] {
				fields["itinerary"] = "No repitas una atracción dentro del mismo día"
			}
			seenAttractions[attractionID] = true
		}
	}
	slices.SortFunc(input.Itinerary, func(a, b packageDayInput) int { return a.DayNumber - b.DayNumber })
	validatePackageSchedule(input.Schedule, fields, create, input.Published, time.Now())
	if input.Published {
		if len([]rune(input.Description)) < 20 {
			fields["published"] = "Para publicar, escribe una descripción de al menos 20 caracteres"
		}
		if input.NationalPriceCents < 1 {
			fields["published"] = "Para publicar, define el precio por persona"
		}
		if len(input.Photos) == 0 {
			fields["published"] = "Para publicar, agrega una fotografía"
		}
		if len(input.Itinerary) == 0 {
			fields["published"] = "Para publicar, agrega al menos un día de itinerario"
		}
	}
	return fields
}

func normalizePackageList(values []string, field string, fields map[string]string) []string {
	if values == nil {
		return []string{}
	}
	if len(values) > 20 {
		fields[field] = "Admite hasta 20 elementos"
	}
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value == "" {
			continue
		}
		if len([]rune(value)) > 200 || seen[key] {
			fields[field] = "Evita elementos repetidos o mayores a 200 caracteres"
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}

func packageError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "No se pudo completar la operación del paquete"
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		status, code, message = 404, "PACKAGE_NOT_FOUND", "El paquete no está disponible o no tienes acceso"
	case errors.Is(err, errPackageConflict):
		status, code, message = 409, "PACKAGE_CONFLICT", "El paquete cambió desde que lo abriste. Recarga antes de guardar"
	case errors.Is(err, errPackageAgencyInactive):
		status, code, message = 403, "PACKAGE_AGENCY_INACTIVE", "La agencia está desactivada. Contacta al administrador"
	case errors.Is(err, errPackagePhoto):
		status, code, message = 422, "PACKAGE_PHOTO_INVALID", "Una fotografía no pertenece a este paquete"
	case errors.Is(err, errPackageAttraction):
		status, code, message = 422, "PACKAGE_ATTRACTION_INVALID", "Selecciona únicamente atracciones activas y publicadas"
	}
	respond.Error(w, r, status, code, message, nil)
}
