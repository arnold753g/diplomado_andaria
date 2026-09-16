package handlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/jpeg"
	"math"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
)

var attractionCategories = []string{"Enoturismo", "Cultural", "Natural", "Deportivo"}
var attractionTimePattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
var errAttractionConflict = errors.New("attraction conflict")
var errAttractionManager = errors.New("attraction manager invalid")
var errAttractionInactive = errors.New("attraction inactive")
var errAttractionPhoto = errors.New("photo does not belong to attraction")

type AttractionHandler struct{ db *gorm.DB }

func NewAttractionHandler(db *gorm.DB) *AttractionHandler { return &AttractionHandler{db} }

type attractionInput struct {
	Name             string   `json:"name" validate:"required,min=2,max=160"`
	Description      string   `json:"description" validate:"max=6000"`
	SubcategoryIDs   []uint64 `json:"subcategory_ids"`
	ScheduleMode     string   `json:"schedule_mode"`
	OpeningTime      string   `json:"opening_time"`
	ClosingTime      string   `json:"closing_time"`
	OpeningDays      []int    `json:"opening_days"`
	SeasonMode       string   `json:"season_mode"`
	SeasonStartMonth *int     `json:"season_start_month"`
	SeasonEndMonth   *int     `json:"season_end_month"`
	Department       string   `json:"department"`
	City             string   `json:"city" validate:"required,min=2,max=100"`
	Address          string   `json:"address" validate:"required,min=3,max=250"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	AdmissionCents   int      `json:"admission_cents" validate:"min=0,max=100000000"`
	Recommendations  string   `json:"recommendations" validate:"max=3000"`
	Phone            string   `json:"phone" validate:"max=30"`
	ManagerID        *uint64  `json:"manager_id"`
	Status           *string  `json:"status"`
	Published        bool     `json:"published"`
	Version          int      `json:"version"`
	Photos           []struct {
		ID   uint64 `json:"id"`
		Data string `json:"data"`
	} `json:"photos"`
}

type attractionAdminView struct {
	models.Attraction
	ManagerID     uint64 `json:"manager_id"`
	ManagerName   string `json:"manager_name"`
	ManagerEmail  string `json:"manager_email"`
	ManagerStatus string `json:"manager_status"`
	Status        string `json:"status"`
	Published     bool   `json:"published"`
	Version       int    `json:"version"`
}

func attractionView(a models.Attraction, scope string) any {
	if scope == "public" {
		return a
	}
	v := attractionAdminView{Attraction: a, ManagerID: a.ManagerID, Status: a.Status, Published: a.Published, Version: a.Version}
	if a.Manager != nil {
		v.ManagerName = strings.TrimSpace(a.Manager.FirstName + " " + a.Manager.LastName)
		v.ManagerEmail = a.Manager.Email
		v.ManagerStatus = a.Manager.Status
	}
	return v
}
func attractionPreload(q *gorm.DB, scope string) *gorm.DB {
	q = q.Preload("Subcategories", func(tx *gorm.DB) *gorm.DB { return tx.Order("position") }).Preload("Subcategories.Subcategory")
	q = q.Preload("Photos", func(tx *gorm.DB) *gorm.DB { return tx.Omit("image").Order("position, id") })
	if scope != "public" {
		q = q.Preload("Manager", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "first_name", "last_name", "email", "status") })
	}
	return q
}
func attractionScope(q *gorm.DB, r *http.Request, scope string) *gorm.DB {
	if scope == "public" {
		return q.Where("attractions.status = ? AND attractions.published = TRUE", models.StatusActive)
	}
	if scope == "mine" {
		p, _ := middleware.CurrentPrincipal(r)
		return q.Where("attractions.manager_id = ?", p.User.ID)
	}
	return q
}
func (h *AttractionHandler) Options(w http.ResponseWriter, r *http.Request) {
	var categories []models.AttractionCategory
	if err := h.db.Preload("Subcategories", func(tx *gorm.DB) *gorm.DB { return tx.Order("position, id") }).Order("position, id").Find(&categories).Error; err != nil {
		attractionError(w, r, err)
		return
	}
	respond.JSON(w, 200, map[string]any{"departments": departments, "categories": categories}, "")
}
func (h *AttractionHandler) Managers(w http.ResponseWriter, r *http.Request) {
	var users []struct {
		ID        uint64 `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	q := h.db.Model(&models.User{}).Select("id, first_name, last_name, email").Where("role = ? AND status = ?", models.RoleAttraction, models.StatusActive)
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		q = q.Where("LOWER(first_name || ' ' || last_name || ' ' || email) LIKE ? ESCAPE '!'", "%"+escapeLike(strings.ToLower(search))+"%")
	}
	if err := q.Order("first_name, last_name, id").Limit(50).Scan(&users).Error; err != nil {
		attractionError(w, r, err)
		return
	}
	respond.JSON(w, 200, users, "")
}
func (h *AttractionHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, "public", false)
}
func (h *AttractionHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, "admin", false)
}
func (h *AttractionHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, "mine", false)
}
func (h *AttractionHandler) Favorites(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, "public", true)
}
func (h *AttractionHandler) list(w http.ResponseWriter, r *http.Request, scope string, favorites bool) {
	page, limit := queryInt(r, "page", 1, 1, 100000), queryInt(r, "limit", 12, 1, 50)
	q := attractionScope(h.db.Model(&models.Attraction{}), r, scope)
	if favorites {
		p, _ := middleware.CurrentPrincipal(r)
		q = q.Where("EXISTS (SELECT 1 FROM attraction_favorites f WHERE f.attraction_id = attractions.id AND f.user_id = ?)", p.User.ID)
	}
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		q = q.Where("LOWER(name || ' ' || city) LIKE ? ESCAPE '!'", "%"+escapeLike(strings.ToLower(search))+"%")
	}
	for _, filter := range []struct {
		key    string
		values []string
	}{{"category", attractionCategories}, {"department", departments}} {
		if value := r.URL.Query().Get(filter.key); value != "" {
			if !slices.Contains(filter.values, value) {
				respond.Error(w, r, 422, "ATTRACTION_VALIDATION", "Filtro inválido", nil)
				return
			}
			if filter.key == "category" {
				q = q.Where("EXISTS (SELECT 1 FROM attraction_classifications ac JOIN attraction_subcategories s ON s.id = ac.subcategory_id JOIN attraction_categories c ON c.id = s.category_id WHERE ac.attraction_id = attractions.id AND c.name = ?)", value)
			} else {
				q = q.Where(filter.key+" = ?", value)
			}
		}
	}
	if value := r.URL.Query().Get("subcategory_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil || id == 0 {
			respond.Error(w, r, 422, "ATTRACTION_VALIDATION", "Subcategoría inválida", nil)
			return
		}
		q = q.Where("EXISTS (SELECT 1 FROM attraction_classifications ac WHERE ac.attraction_id = attractions.id AND ac.subcategory_id = ?)", id)
	}
	if scope != "public" {
		if status := r.URL.Query().Get("status"); status != "" {
			if !validStatus(status) {
				respond.Error(w, r, 422, "ATTRACTION_VALIDATION", "Estado inválido", nil)
				return
			}
			q = q.Where("status = ?", status)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		attractionError(w, r, err)
		return
	}
	var attractions []models.Attraction
	if err := attractionPreload(q, scope).Order("attractions.created_at DESC, attractions.id DESC").Limit(limit).Offset((page - 1) * limit).Find(&attractions).Error; err != nil {
		attractionError(w, r, err)
		return
	}
	views := make([]any, 0, len(attractions))
	for _, a := range attractions {
		views = append(views, attractionView(a, scope))
	}
	respond.JSON(w, 200, map[string]any{"attractions": views, "pagination": map[string]any{"page": page, "limit": limit, "total": total}}, "")
}
func (h *AttractionHandler) GetPublic(w http.ResponseWriter, r *http.Request) { h.get(w, r, "public") }
func (h *AttractionHandler) GetAdmin(w http.ResponseWriter, r *http.Request)  { h.get(w, r, "admin") }
func (h *AttractionHandler) GetMine(w http.ResponseWriter, r *http.Request)   { h.get(w, r, "mine") }
func (h *AttractionHandler) get(w http.ResponseWriter, r *http.Request, scope string) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var a models.Attraction
	if err := attractionPreload(attractionScope(h.db, r, scope), scope).First(&a, id).Error; err != nil {
		attractionError(w, r, err)
		return
	}
	respond.JSON(w, 200, attractionView(a, scope), "")
}
func (h *AttractionHandler) PhotoPublic(w http.ResponseWriter, r *http.Request) {
	h.photo(w, r, "public")
}
func (h *AttractionHandler) PhotoAdmin(w http.ResponseWriter, r *http.Request) {
	h.photo(w, r, "admin")
}
func (h *AttractionHandler) PhotoMine(w http.ResponseWriter, r *http.Request) { h.photo(w, r, "mine") }
func (h *AttractionHandler) photo(w http.ResponseWriter, r *http.Request, scope string) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	photoID, err := strconv.ParseUint(mux.Vars(r)["photo"], 10, 64)
	if err != nil {
		attractionError(w, r, gorm.ErrRecordNotFound)
		return
	}
	var photo models.AttractionPhoto
	// One query checks visibility and photo ownership, including unpublished images.
	permitted := attractionScope(h.db.Model(&models.Attraction{}).Select("id"), r, scope).Where("id = ?", id)
	if err := h.db.Where("id = ? AND attraction_id IN (?)", photoID, permitted).First(&photo).Error; err != nil {
		attractionError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(200)
	_, _ = w.Write(photo.Image)
}
func (h *AttractionHandler) Create(w http.ResponseWriter, r *http.Request) { h.save(w, r, true, false) }
func (h *AttractionHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	h.save(w, r, false, false)
}
func (h *AttractionHandler) UpdateMine(w http.ResponseWriter, r *http.Request) {
	h.save(w, r, false, true)
}
func (h *AttractionHandler) save(w http.ResponseWriter, r *http.Request, create, mine bool) {
	var input attractionInput
	if !decodeJSONLimit(w, r, &input, 42<<20) {
		return
	}
	if mine && (input.ManagerID != nil || input.Status != nil) {
		respond.Error(w, r, 403, "ATTRACTION_ADMIN_ONLY", "Solo el administrador asigna encargados o cambia el estado", nil)
		return
	}
	for _, s := range []*string{&input.Name, &input.Description, &input.City, &input.Address, &input.Recommendations, &input.Phone} {
		*s = strings.TrimSpace(*s)
	}
	fields := map[string]string{}
	if err := validator.New().Struct(input); err != nil {
		fields["data"] = "Revisa los campos obligatorios, sus longitudes y el precio de entrada"
	}
	validateAttractionVisit(&input, fields)
	var selected []models.AttractionSubcategory
	if len(input.SubcategoryIDs) >= 1 && len(input.SubcategoryIDs) <= 4 {
		if err := h.db.Where("id IN ?", input.SubcategoryIDs).Find(&selected).Error; err != nil {
			attractionError(w, r, err)
			return
		}
		if len(selected) != len(input.SubcategoryIDs) {
			fields["subcategory_ids"] = "Selecciona subcategorías válidas y sin repetir"
		}
	} else {
		fields["subcategory_ids"] = "Selecciona entre una y cuatro subcategorías"
	}
	primaryCategory := ""
	if len(fields) == 0 {
		if err := h.db.Model(&models.AttractionCategory{}).Select("name").Where("id = (SELECT category_id FROM attraction_subcategories WHERE id = ?)", input.SubcategoryIDs[0]).Scan(&primaryCategory).Error; err != nil {
			attractionError(w, r, err)
			return
		}
	}
	if !slices.Contains(departments, input.Department) {
		fields["department"] = "Selecciona un departamento de Bolivia"
	}
	if (input.Latitude == nil) != (input.Longitude == nil) {
		fields["coordinates"] = "Completa ambas coordenadas o deja ambas vacías"
	}
	if input.Latitude != nil && (math.IsNaN(*input.Latitude) || math.IsInf(*input.Latitude, 0) || *input.Latitude < -90 || *input.Latitude > 90) {
		fields["latitude"] = "Latitud entre -90 y 90"
	}
	if input.Longitude != nil && (math.IsNaN(*input.Longitude) || math.IsInf(*input.Longitude, 0) || *input.Longitude < -180 || *input.Longitude > 180) {
		fields["longitude"] = "Longitud entre -180 y 180"
	}
	if !mine && (input.ManagerID == nil || *input.ManagerID == 0) {
		fields["manager_id"] = "Selecciona un encargado de atracción"
	}
	if input.Status != nil && !validStatus(*input.Status) {
		fields["status"] = "Estado inválido"
	}
	if !create && input.Version < 1 {
		fields["version"] = "Recarga la atracción antes de guardar"
	}
	if len(input.Photos) > 6 {
		fields["photos"] = "Puedes guardar hasta seis fotografías"
	}
	if input.Published && (len([]rune(input.Description)) < 20 || len(input.Photos) == 0) {
		fields["published"] = "Para publicar, completa una descripción de al menos 20 caracteres y una fotografía"
	}
	images := make(map[int][]byte)
	seen := make(map[uint64]bool)
	if len(input.Photos) <= 6 {
		for i, photo := range input.Photos {
			if photo.ID > 0 {
				if create || seen[photo.ID] || photo.Data != "" {
					fields["photos"] = "Las fotografías existentes no son válidas"
				}
				seen[photo.ID] = true
				continue
			}
			img, err := decodeAttractionImage(photo.Data)
			if err != nil {
				fields["photos"] = err.Error()
				break
			}
			images[i] = img
		}
	}
	if len(fields) > 0 {
		respond.Error(w, r, 422, "ATTRACTION_VALIDATION", "Revisa los datos de la atracción", fields)
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
	var a models.Attraction
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(7412001)").Error; err != nil {
			return err
		}
		if create {
			a.Status = models.StatusActive
			a.Version = 1
		} else {
			q := tx.Clauses(clause.Locking{Strength: "UPDATE"})
			if mine {
				p, _ := middleware.CurrentPrincipal(r)
				q = q.Where("manager_id = ?", p.User.ID)
			}
			if err := q.First(&a, id).Error; err != nil {
				return err
			}
			if mine && a.Status != models.StatusActive {
				return errAttractionInactive
			}
			if a.Version != input.Version {
				return errAttractionConflict
			}
			a.Version++
		}
		if !mine {
			var manager models.User
			if err := tx.Where("id = ? AND role = ?", *input.ManagerID, models.RoleAttraction).First(&manager).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errAttractionManager
				}
				return err
			}
			if manager.Status != models.StatusActive && (create || a.ManagerID != manager.ID) {
				return errAttractionManager
			}
			a.ManagerID = manager.ID
			if input.Status != nil {
				a.Status = *input.Status
			}
		}
		a.Name = input.Name
		a.Description = input.Description
		a.Category = primaryCategory
		a.Department = input.Department
		a.City = input.City
		a.Address = input.Address
		a.Latitude = input.Latitude
		a.Longitude = input.Longitude
		a.ScheduleMode = input.ScheduleMode
		a.OpeningTime = input.OpeningTime
		a.ClosingTime = input.ClosingTime
		a.OpeningDays = input.OpeningDays
		a.SeasonMode = input.SeasonMode
		a.SeasonStartMonth = input.SeasonStartMonth
		a.SeasonEndMonth = input.SeasonEndMonth
		if input.ScheduleMode != "unspecified" {
			a.OpeningHours = ""
		}
		a.AdmissionCents = input.AdmissionCents
		a.Recommendations = input.Recommendations
		a.Phone = input.Phone
		a.Published = input.Published && a.Status == models.StatusActive
		if create {
			if err := tx.Create(&a).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(&a).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("attraction_id = ?", a.ID).Delete(&models.AttractionClassification{}).Error; err != nil {
			return err
		}
		for i, id := range input.SubcategoryIDs {
			if err := tx.Create(&models.AttractionClassification{AttractionID: a.ID, SubcategoryID: id, Position: i}).Error; err != nil {
				return err
			}
		}
		keep := make([]uint64, 0, len(input.Photos))
		for i, photo := range input.Photos {
			if photo.ID > 0 {
				res := tx.Model(&models.AttractionPhoto{}).Where("id = ? AND attraction_id = ?", photo.ID, a.ID).Update("position", i)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected != 1 {
					return errAttractionPhoto
				}
				keep = append(keep, photo.ID)
			} else {
				photo := models.AttractionPhoto{AttractionID: a.ID, Position: i, Image: images[i]}
				if err := tx.Create(&photo).Error; err != nil {
					return err
				}
				keep = append(keep, photo.ID)
			}
		}
		q := tx.Where("attraction_id = ?", a.ID)
		if len(keep) > 0 {
			q = q.Where("id NOT IN ?", keep)
		}
		return q.Delete(&models.AttractionPhoto{}).Error
	})
	if err != nil {
		attractionError(w, r, err)
		return
	}
	if err := attractionPreload(h.db, "admin").First(&a, a.ID).Error; err != nil {
		attractionError(w, r, err)
		return
	}
	status := 200
	if create {
		status = 201
	}
	respond.JSON(w, status, attractionView(a, "admin"), "Atracción guardada")
}

func decodeAttractionImage(data string) ([]byte, error) {
	invalid := errors.New("Usa fotos PNG o JPG de hasta 5 MB y 4096 × 4096 píxeles")
	prefix, encoded, ok := strings.Cut(data, ",")
	if !ok || (prefix != "data:image/png;base64" && prefix != "data:image/jpeg;base64") || len(encoded) > base64.StdEncoding.EncodedLen(5<<20) {
		return nil, invalid
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) > 5<<20 {
		return nil, invalid
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "png" && format != "jpeg") || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 4096 || cfg.Height > 4096 {
		return nil, invalid
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, invalid
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, invalid
	}
	if out.Len() > 5<<20 {
		return nil, errors.New("Reduce la resolución de la foto: la imagen procesada supera 5 MB")
	}
	return out.Bytes(), nil
}

func (h *AttractionHandler) Favorite(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	p, _ := middleware.CurrentPrincipal(r)
	if r.Method == http.MethodDelete {
		if err := h.db.Where("user_id = ? AND attraction_id = ?", p.User.ID, id).Delete(&models.AttractionFavorite{}).Error; err != nil {
			attractionError(w, r, err)
			return
		}
		respond.JSON(w, 200, map[string]bool{"favorite": false}, "")
		return
	}
	favorite := false
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var a models.Attraction
		if err := attractionScope(tx, r, "public").Clauses(clause.Locking{Strength: "SHARE"}).First(&a, id).Error; err != nil {
			return err
		}
		if r.Method == http.MethodPut {
			favorite = true
			return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.AttractionFavorite{UserID: p.User.ID, AttractionID: id}).Error
		}
		var count int64
		if err := tx.Model(&models.AttractionFavorite{}).Where("user_id = ? AND attraction_id = ?", p.User.ID, id).Count(&count).Error; err != nil {
			return err
		}
		favorite = count > 0
		return nil
	})
	if err != nil {
		attractionError(w, r, err)
		return
	}
	respond.JSON(w, 200, map[string]bool{"favorite": favorite}, "")
}
func attractionError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := 500, "INTERNAL_ERROR", "No se pudo completar la operación de atracción"
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		status, code, message = 404, "ATTRACTION_NOT_FOUND", "La atracción no está disponible o no tienes acceso"
	case errors.Is(err, errAttractionConflict):
		status, code, message = 409, "ATTRACTION_CONFLICT", "La atracción cambió desde que la abriste. Recarga antes de guardar"
	case errors.Is(err, errAttractionManager):
		status, code, message = 422, "ATTRACTION_MANAGER_INVALID", "Selecciona un encargado de atracción activo"
	case errors.Is(err, errAttractionInactive):
		status, code, message = 403, "ATTRACTION_INACTIVE", "La atracción está desactivada. Contacta al administrador"
	case errors.Is(err, errAttractionPhoto):
		status, code, message = 422, "ATTRACTION_PHOTO_INVALID", "Una fotografía no pertenece a esta atracción"
	}
	respond.Error(w, r, status, code, message, nil)
}

// Hours are local wall-clock values, independent of the browser's timezone.
func validateAttractionVisit(input *attractionInput, fields map[string]string) {
	if input.ScheduleMode == "" {
		input.ScheduleMode = "unspecified"
	}
	if input.SeasonMode == "" {
		input.SeasonMode = "unspecified"
	}
	if input.OpeningDays == nil {
		input.OpeningDays = []int{}
	}
	if !slices.Contains([]string{"unspecified", "scheduled", "all_day"}, input.ScheduleMode) {
		fields["schedule_mode"] = "Selecciona un tipo de horario válido"
	}
	if input.ScheduleMode == "unspecified" {
		if input.OpeningTime != "" || input.ClosingTime != "" || len(input.OpeningDays) != 0 {
			fields["schedule_mode"] = "Selecciona un horario para guardar horas o días"
		}
	} else {
		if len(input.OpeningDays) < 1 || len(input.OpeningDays) > 7 {
			fields["opening_days"] = "Selecciona de uno a siete días de apertura"
		}
		seen := map[int]bool{}
		for _, day := range input.OpeningDays {
			if day < 1 || day > 7 || seen[day] {
				fields["opening_days"] = "Los días deben ser válidos y no repetirse"
			}
			seen[day] = true
		}
		slices.Sort(input.OpeningDays)
		if input.ScheduleMode == "scheduled" {
			if !attractionTimePattern.MatchString(input.OpeningTime) || !attractionTimePattern.MatchString(input.ClosingTime) || input.OpeningTime == input.ClosingTime {
				fields["opening_time"] = "Selecciona horas válidas y distintas; usa 24 horas si abre todo el día"
			}
		} else if input.OpeningTime != "" || input.ClosingTime != "" {
			fields["opening_time"] = "El horario de 24 horas no lleva horas de apertura o cierre"
		}
	}
	if !slices.Contains([]string{"unspecified", "all_year", "months"}, input.SeasonMode) {
		fields["season_mode"] = "Selecciona una temporada válida"
	}
	if input.SeasonMode == "months" {
		if input.SeasonStartMonth == nil || input.SeasonEndMonth == nil || *input.SeasonStartMonth < 1 || *input.SeasonStartMonth > 12 || *input.SeasonEndMonth < 1 || *input.SeasonEndMonth > 12 {
			fields["season_months"] = "Selecciona los meses de inicio y fin"
		}
	} else if input.SeasonStartMonth != nil || input.SeasonEndMonth != nil {
		fields["season_months"] = "Los meses solo se usan para una temporada específica"
	}
}
