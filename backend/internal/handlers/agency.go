package handlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
)

var departments = []string{"Tarija", "Chuquisaca", "La Paz", "Cochabamba", "Oruro", "Potosí", "Santa Cruz", "Beni", "Pando"}
var errAgencyConflict = errors.New("agency version conflict")
var errAgencyManager = errors.New("invalid agency manager")
var errAgencyInactive = errors.New("agency inactive")
var errAgencyAssignment = errors.New("manager already assigned")

type AgencyHandler struct{ db *gorm.DB }

func NewAgencyHandler(db *gorm.DB) *AgencyHandler { return &AgencyHandler{db: db} }

type agencyRequest struct {
	Name                string  `json:"name" validate:"required,min=2,max=160"`
	Description         string  `json:"description" validate:"max=3000"`
	Department          string  `json:"department" validate:"required"`
	City                string  `json:"city" validate:"required,min=2,max=100"`
	Address             string  `json:"address" validate:"required,min=3,max=250"`
	Phone               string  `json:"phone" validate:"required,min=7,max=30"`
	Email               string  `json:"email" validate:"required,email,max=320"`
	ManagerID           *uint64 `json:"manager_id,omitempty"`
	Status              *string `json:"status,omitempty"`
	Published           bool    `json:"published"`
	MinimumPayingAge    *int    `json:"minimum_paying_age"`
	AcceptsQR           bool    `json:"accepts_qr"`
	AcceptsTransfer     bool    `json:"accepts_transfer"`
	BankName            string  `json:"bank_name" validate:"max=100"`
	AccountHolder       string  `json:"account_holder" validate:"max=160"`
	AccountNumber       string  `json:"account_number" validate:"max=50"`
	PaymentInstructions string  `json:"payment_instructions" validate:"max=1000"`
	QRImage             string  `json:"qr_image"`
	Version             int     `json:"version"`
}

type agencyResponse struct {
	models.Agency
	QRImage       string `json:"qr_image,omitempty"`
	ManagerName   string `json:"manager_name"`
	ManagerEmail  string `json:"manager_email"`
	ManagerStatus string `json:"manager_status"`
}

func agencyView(a models.Agency, detail bool) agencyResponse {
	v := agencyResponse{Agency: a}
	if a.Manager != nil {
		v.ManagerName = strings.TrimSpace(a.Manager.FirstName + " " + a.Manager.LastName)
		v.ManagerEmail = a.Manager.Email
		v.ManagerStatus = a.Manager.Status
	}
	if detail && len(a.QRImage) > 0 {
		v.QRImage = "data:image/png;base64," + base64.StdEncoding.EncodeToString(a.QRImage)
	}
	return v
}
func agencyQuery(db *gorm.DB) *gorm.DB {
	return db.Preload("Manager", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "first_name", "last_name", "email", "status") })
}

func (h *AgencyHandler) Options(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, 200, map[string]any{"departments": departments}, "")
}

// Only IDs/names/emails needed for assignment; no account documents or passwords.
func (h *AgencyHandler) Managers(w http.ResponseWriter, r *http.Request) {
	var users []struct {
		ID        uint64 `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	q := h.db.Model(&models.User{}).Select("id, first_name, last_name, email").Where("role = ? AND status = ?", models.RoleAgency, models.StatusActive).
		Where("NOT EXISTS (SELECT 1 FROM agencies WHERE agencies.manager_id = users.id)")
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		pattern := "%" + escapeLike(strings.ToLower(search)) + "%"
		q = q.Where("LOWER(first_name || ' ' || last_name || ' ' || email) LIKE ? ESCAPE '!'", pattern)
	}
	if err := q.Order("first_name, last_name, id").Limit(50).Scan(&users).Error; err != nil {
		agencyError(w, r, err)
		return
	}
	respond.JSON(w, 200, users, "")
}

func (h *AgencyHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := queryInt(r, "page", 1, 1, 100000), queryInt(r, "limit", 12, 1, 50)
	q := h.db.Model(&models.Agency{})
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		q = q.Where("LOWER(name || ' ' || city) LIKE ? ESCAPE '!'", "%"+escapeLike(strings.ToLower(search))+"%")
	}
	if status := r.URL.Query().Get("status"); status != "" {
		if !validStatus(status) {
			respond.Error(w, r, 422, "AGENCY_VALIDATION", "Selecciona un estado válido", nil)
			return
		}
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		agencyError(w, r, err)
		return
	}
	var agencies []models.Agency
	if err := agencyQuery(q).Omit("qr_image").Order("created_at DESC, id DESC").Limit(limit).Offset((page - 1) * limit).Find(&agencies).Error; err != nil {
		agencyError(w, r, err)
		return
	}
	views := make([]agencyResponse, 0, len(agencies))
	for _, a := range agencies {
		views = append(views, agencyView(a, false))
	}
	respond.JSON(w, 200, map[string]any{"agencies": views, "pagination": map[string]any{"page": page, "limit": limit, "total": total}}, "")
}

func (h *AgencyHandler) Get(w http.ResponseWriter, r *http.Request)  { h.get(w, r, false) }
func (h *AgencyHandler) Mine(w http.ResponseWriter, r *http.Request) { h.get(w, r, true) }
func (h *AgencyHandler) get(w http.ResponseWriter, r *http.Request, mine bool) {
	q := agencyQuery(h.db)
	if mine {
		p, _ := middleware.CurrentPrincipal(r)
		q = q.Where("manager_id = ?", p.User.ID)
	} else {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		q = q.Where("id = ?", id)
	}
	var a models.Agency
	if err := q.First(&a).Error; err != nil {
		agencyError(w, r, err)
		return
	}
	respond.JSON(w, 200, agencyView(a, true), "")
}
func (h *AgencyHandler) Create(w http.ResponseWriter, r *http.Request)     { h.save(w, r, true, false) }
func (h *AgencyHandler) Update(w http.ResponseWriter, r *http.Request)     { h.save(w, r, false, false) }
func (h *AgencyHandler) UpdateMine(w http.ResponseWriter, r *http.Request) { h.save(w, r, false, true) }

func (h *AgencyHandler) save(w http.ResponseWriter, r *http.Request, create, mine bool) {
	p, _ := middleware.CurrentPrincipal(r)
	var input agencyRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if mine && (input.ManagerID != nil || input.Status != nil) {
		respond.Error(w, r, 403, "AGENCY_ADMIN_ONLY", "Solo el administrador cambia el encargado o el estado de la agencia", nil)
		return
	}
	for _, s := range []*string{&input.Name, &input.Description, &input.Department, &input.City, &input.Address, &input.Phone, &input.BankName, &input.AccountHolder, &input.AccountNumber, &input.PaymentInstructions} {
		*s = strings.TrimSpace(*s)
	}
	input.Email = normalizeEmail(input.Email)
	if input.MinimumPayingAge == nil && create {
		age := 6
		input.MinimumPayingAge = &age
	}
	fields := map[string]string{}
	if err := validator.New().Struct(input); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, e := range ve {
				fields[e.Field()] = "Revisa este dato y su longitud"
			}
		}
	}
	validDept := false
	for _, d := range departments {
		if input.Department == d {
			validDept = true
		}
	}
	if !validDept {
		fields["department"] = "Selecciona un departamento de Bolivia"
	}
	if input.MinimumPayingAge == nil || *input.MinimumPayingAge < 0 || *input.MinimumPayingAge > 18 {
		fields["minimum_paying_age"] = "Ingresa una edad entre 0 y 18"
	}
	if input.AcceptsTransfer && (input.BankName == "" || input.AccountHolder == "" || input.AccountNumber == "") {
		fields["transfer"] = "Completa banco, titular y número de cuenta"
	}
	if !mine && (input.ManagerID == nil || *input.ManagerID == 0) {
		fields["manager_id"] = "Selecciona un encargado"
	}
	if input.Status != nil && !validStatus(*input.Status) {
		fields["status"] = "Estado inválido"
	}
	if !create && input.Version < 1 {
		fields["version"] = "Recarga la agencia antes de guardar"
	}
	qr, err := decodeAgencyQR(input.QRImage)
	if err != nil {
		fields["qr_image"] = err.Error()
	}
	if input.AcceptsQR && len(qr) == 0 {
		fields["qr_image"] = "Sube el QR para habilitar este medio de pago"
	}
	if len(fields) > 0 {
		respond.Error(w, r, 422, "AGENCY_VALIDATION", "Revisa los datos de la agencia", fields)
		return
	}
	var agency models.Agency
	var id uint64
	if !create && !mine {
		var ok bool
		id, ok = pathID(w, r)
		if !ok {
			return
		}
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Same lock as user-role mutations: assignment and role changes cannot race.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(7412001)").Error; err != nil {
			return err
		}
		if create {
			agency.Status = models.StatusActive
			agency.Version = 1
		} else {
			q := tx.Clauses(clause.Locking{Strength: "UPDATE"})
			if mine {
				q = q.Where("manager_id = ?", p.User.ID)
			} else {
				q = q.Where("id = ?", id)
			}
			if err := q.First(&agency).Error; err != nil {
				return err
			}
			if mine && agency.Status != models.StatusActive {
				return errAgencyInactive
			}
			if agency.Version != input.Version {
				return errAgencyConflict
			}
			agency.Version++
		}
		if !mine {
			var manager models.User
			if err := tx.Where("id = ? AND role = ? AND status = ?", *input.ManagerID, models.RoleAgency, models.StatusActive).First(&manager).Error; err != nil {
				// An admin may deactivate an agency whose current manager is inactive.
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if create || agency.ManagerID != *input.ManagerID {
					return errAgencyManager
				}
				if err := tx.Where("id = ? AND role = ?", *input.ManagerID, models.RoleAgency).First(&manager).Error; err != nil {
					return errAgencyManager
				}
			}
			var assigned int64
			if err := tx.Model(&models.Agency{}).Where("manager_id = ? AND id <> ?", *input.ManagerID, agency.ID).Count(&assigned).Error; err != nil {
				return err
			}
			if assigned > 0 {
				return errAgencyAssignment
			}
			agency.ManagerID = *input.ManagerID
			if input.Status != nil {
				agency.Status = *input.Status
			}
		}
		agency.Name = input.Name
		agency.Description = input.Description
		agency.Department = input.Department
		agency.City = input.City
		agency.Address = input.Address
		agency.Phone = input.Phone
		agency.Email = input.Email
		agency.Published = input.Published
		if agency.Status == models.StatusInactive {
			agency.Published = false
		}
		agency.MinimumPayingAge = *input.MinimumPayingAge
		agency.AcceptsQR = input.AcceptsQR
		agency.AcceptsTransfer = input.AcceptsTransfer
		agency.BankName = input.BankName
		agency.AccountHolder = input.AccountHolder
		agency.AccountNumber = input.AccountNumber
		agency.PaymentInstructions = input.PaymentInstructions
		agency.QRImage = qr
		if create {
			return tx.Create(&agency).Error
		}
		return tx.Save(&agency).Error
	})
	if err != nil {
		agencyError(w, r, err)
		return
	}
	if err := agencyQuery(h.db).First(&agency, agency.ID).Error; err != nil {
		agencyError(w, r, err)
		return
	}
	code := http.StatusOK
	if create {
		code = http.StatusCreated
	}
	respond.JSON(w, code, agencyView(agency, true), "Agencia guardada")
}

func decodeAgencyQR(data string) ([]byte, error) {
	if data == "" {
		return nil, nil
	}
	prefix, encoded, ok := strings.Cut(data, ",")
	if !ok || (prefix != "data:image/png;base64" && prefix != "data:image/jpeg;base64") || len(encoded) > 700000 {
		return nil, errors.New("Usa una imagen PNG o JPG de hasta 500 KB")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) > 500*1024 {
		return nil, errors.New("El archivo QR no es válido o supera 500 KB")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "png" && format != "jpeg") || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 2048 || cfg.Height > 2048 {
		return nil, errors.New("La imagen debe ser PNG o JPG y no superar 2048 × 2048 píxeles")
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, errors.New("No se pudo leer la imagen")
	}
	var output bytes.Buffer
	if err := png.Encode(&output, img); err != nil {
		return nil, err
	}
	// Canonical encoding strips metadata and prevents serving arbitrary uploaded content.
	if output.Len() > 500*1024 {
		return nil, errors.New("Reduce el tamaño de la imagen QR a 500 KB")
	}
	return output.Bytes(), nil
}
func agencyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		respond.Error(w, r, 404, "AGENCY_NOT_FOUND", "No tienes una agencia asignada o la agencia no existe", nil)
	case errors.Is(err, errAgencyConflict):
		respond.Error(w, r, 409, "AGENCY_CONFLICT", "La agencia cambió desde que la abriste. Recarga sus datos antes de guardar", nil)
	case errors.Is(err, errAgencyManager):
		respond.Error(w, r, 422, "AGENCY_MANAGER_INVALID", "Selecciona un usuario activo con rol de encargado de agencia", nil)
	case errors.Is(err, errAgencyAssignment), errors.Is(err, gorm.ErrDuplicatedKey):
		respond.Error(w, r, 409, "AGENCY_MANAGER_ASSIGNED", "Ese encargado ya tiene una agencia asignada", nil)
	case errors.Is(err, errAgencyInactive):
		respond.Error(w, r, 403, "AGENCY_INACTIVE", "La agencia está desactivada. Contacta al administrador", nil)
	default:
		respond.Error(w, r, 500, "INTERNAL_ERROR", "No se pudo completar la operación de agencia", nil)
	}
}
