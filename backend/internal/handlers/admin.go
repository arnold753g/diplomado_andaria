package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type AdminHandler struct{ db *gorm.DB }

func NewAdminHandler(db *gorm.DB) *AdminHandler { return &AdminHandler{db: db} }

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int64{}
	queries := []struct {
		name  string
		query *gorm.DB
	}{
		{"users", h.db.Model(&models.User{})},
		{"active_users", h.db.Model(&models.User{}).Where("status = ?", models.StatusActive)},
		{"active_admins", h.db.Model(&models.User{}).Where("role = ? AND status = ?", models.RoleAdmin, models.StatusActive)},
		{"agencies", h.db.Model(&models.Agency{})},
		{"published_agencies", h.db.Model(&models.Agency{}).Where("status = ? AND published = TRUE", models.StatusActive)},
		{"attractions", h.db.Model(&models.Attraction{})},
		{"published_attractions", h.db.Model(&models.Attraction{}).Where("status = ? AND published = TRUE", models.StatusActive)},
		{"packages", h.db.Model(&models.TourPackage{})},
		{"published_packages", h.db.Model(&models.TourPackage{}).Where("published = TRUE")},
		{"purchases", h.db.Model(&models.TourPackagePurchase{})},
	}
	for _, item := range queries {
		var count int64
		if err := item.query.Count(&count).Error; err != nil {
			respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Dashboard could not be loaded", nil)
			return
		}
		stats[item.name] = count
	}
	respond.JSON(w, http.StatusOK, stats, "Dashboard loaded")
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1, 1, 1000000)
	limit := queryInt(r, "limit", 20, 1, 100)
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	query := h.db.Model(&models.User{})
	if search != "" {
		like := "%" + escapeLike(strings.ToLower(search)) + "%"
		query = query.Where("LOWER(email) LIKE ? ESCAPE '!' OR LOWER(first_name || ' ' || last_name) LIKE ? ESCAPE '!'", like, like)
	}
	if role := strings.TrimSpace(r.URL.Query().Get("role")); role != "" {
		if !validRole(role) {
			respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid role filter", nil)
			return
		}
		query = query.Where("role = ?", role)
	}
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		if !validStatus(status) {
			respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid status filter", nil)
			return
		}
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Users could not be loaded", nil)
		return
	}
	var users []models.User
	if err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error; err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Users could not be loaded", nil)
		return
	}
	public := make([]models.PublicUser, len(users))
	for index, user := range users {
		public[index] = user.Public()
	}
	respond.JSON(w, http.StatusOK, map[string]any{"users": public, "pagination": map[string]any{
		"page": page, "limit": limit, "total": total, "total_pages": (total + int64(limit) - 1) / int64(limit),
	}}, "Users loaded")
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", "User was not found", nil)
			return
		}
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "User could not be loaded", nil)
		return
	}
	respond.JSON(w, http.StatusOK, user.Public(), "User loaded")
}

func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok || !h.canManageTarget(w, r, id) {
		return
	}
	var request struct {
		Role string `json:"role"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if !validRole(request.Role) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Selecciona un rol válido", map[string]string{"role": "Rol inválido"})
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := guardLastAdmin(tx, id, request.Role, ""); err != nil {
			return err
		}
		result := tx.Model(&models.User{}).Where("id = ?", id).Update("role", request.Role)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		now := time.Now().UTC()
		return tx.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
	}); err != nil {
		handleAdminMutationError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"id": id, "role": request.Role}, "Role updated; existing sessions revoked")
}

func (h *AdminHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok || !h.canManageTarget(w, r, id) {
		return
	}
	var request struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if !validStatus(request.Status) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Status must be active or inactive", map[string]string{"status": "Invalid status"})
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := guardLastAdmin(tx, id, "", request.Status); err != nil {
			return err
		}
		result := tx.Model(&models.User{}).Where("id = ?", id).Update("status", request.Status)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if request.Status == models.StatusInactive {
			now := time.Now().UTC()
			return tx.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
		}
		return nil
	}); err != nil {
		handleAdminMutationError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"id": id, "status": request.Status}, "Status updated")
}

func (h *AdminHandler) canManageTarget(w http.ResponseWriter, r *http.Request, id uint64) bool {
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return false
	}
	if principal.User.ID == id {
		respond.Error(w, r, http.StatusConflict, "SELF_ADMIN_CHANGE", "Use profile settings for your own account", nil)
		return false
	}
	return true
}

func handleAdminMutationError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, errManagerHasAttractions) {
		respond.Error(w, r, 409, "MANAGER_HAS_ATTRACTIONS", "Reasigna sus atracciones antes de cambiar el rol de este encargado", nil)
		return
	}
	if errors.Is(err, errManagerHasAgency) {
		respond.Error(w, r, 409, "MANAGER_HAS_AGENCY", "Reasigna su agencia antes de cambiar el rol de este encargado", nil)
		return
	}
	if errors.Is(err, errLastAdmin) {
		respond.Error(w, r, 409, "LAST_ADMIN", "Debe existir al menos un administrador activo", nil)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", "User was not found", nil)
		return
	}
	respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "User could not be updated", nil)
}

func pathID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		respond.Error(w, r, http.StatusBadRequest, "INVALID_ID", "User ID is invalid", nil)
		return 0, false
	}
	return id, true
}

func queryInt(r *http.Request, key string, fallback, min, max int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || value < min {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	return replacer.Replace(value)
}

var errLastAdmin = errors.New("last active admin")
var errManagerHasAttractions = errors.New("manager has attractions")
var errManagerHasAgency = errors.New("manager has agency")

func guardLastAdmin(tx *gorm.DB, id uint64, role, status string) error {
	if err := tx.Exec("SELECT pg_advisory_xact_lock(7412001)").Error; err != nil {
		return err
	}
	var user models.User
	if err := tx.First(&user, id).Error; err != nil {
		return err
	}
	if role == "" {
		role = user.Role
	}
	if status == "" {
		status = user.Status
	}
	if user.Role == models.RoleAttraction && role != models.RoleAttraction {
		var assigned int64
		if err := tx.Model(&models.Attraction{}).Where("manager_id = ?", id).Count(&assigned).Error; err != nil {
			return err
		}
		if assigned > 0 {
			return errManagerHasAttractions
		}
	}
	if user.Role == models.RoleAgency && role != models.RoleAgency {
		var assigned int64
		if err := tx.Model(&models.Agency{}).Where("manager_id = ?", id).Count(&assigned).Error; err != nil {
			return err
		}
		if assigned > 0 {
			return errManagerHasAgency
		}
	}
	if user.Role == models.RoleAdmin && user.Status == models.StatusActive && (role != models.RoleAdmin || status != models.StatusActive) {
		var count int64
		if err := tx.Model(&models.User{}).Where("role = ? AND status = ?", models.RoleAdmin, models.StatusActive).Count(&count).Error; err != nil {
			return err
		}
		if count <= 1 {
			return errLastAdmin
		}
	}
	return nil
}

// Edit the user in one transaction so profile, role and status never partially save.
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok || !h.canManageTarget(w, r, id) {
		return
	}
	var request struct {
		updateProfileRequest
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	if err := validator.New().Struct(request); err != nil || !validRole(request.Role) || !validStatus(request.Status) {
		respond.Error(w, r, 422, "VALIDATION_ERROR", "Revisa los datos, rol y estado", nil)
		return
	}
	var updated models.User
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := guardLastAdmin(tx, id, request.Role, request.Status); err != nil {
			return err
		}
		var original models.User
		if err := tx.First(&original, id).Error; err != nil {
			return err
		}
		revoke := original.Role != request.Role || request.Status == models.StatusInactive
		if err := tx.Model(&original).Updates(map[string]any{"first_name": request.FirstName, "last_name": request.LastName, "phone": strings.TrimSpace(request.Phone), "document_number": strings.TrimSpace(request.DocumentNumber), "nationality": strings.TrimSpace(request.Nationality), "role": request.Role, "status": request.Status}).Error; err != nil {
			return err
		}
		if revoke {
			if err := middleware.RevokeUserSessions(tx, id); err != nil {
				return err
			}
		}
		return tx.First(&updated, id).Error
	})
	if err != nil {
		handleAdminMutationError(w, r, err)
		return
	}
	respond.JSON(w, 200, updated.Public(), "Usuario actualizado")
}
