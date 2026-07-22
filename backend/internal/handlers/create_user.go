package handlers

import (
	"errors"
	"gorm.io/gorm"
	"net/http"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	"starter-backend/internal/security"
	"strings"
)

// Only mounted beneath the admin middleware. Public registration never accepts a role.
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request struct {
		registerRequest
		Role string `json:"role"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Email = normalizeEmail(request.Email)
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	if err := h.validate.Struct(request); err != nil || !models.ValidRole(request.Role) {
		respond.Error(w, r, 422, "VALIDATION_ERROR", "Revisa los datos y el rol del usuario", nil)
		return
	}
	hash, err := security.HashPassword(request.Password, h.cfg.PasswordBcryptCost)
	if err != nil {
		respond.Error(w, r, 422, "PASSWORD_POLICY", "Usa una contraseña de 12 a 72 bytes, sin espacios en los extremos", nil)
		return
	}
	user := models.User{Email: request.Email, PasswordHash: hash, FirstName: request.FirstName, LastName: request.LastName, Role: request.Role, Status: models.StatusActive,
		Phone: strings.TrimSpace(request.Phone), DocumentNumber: strings.TrimSpace(request.DocumentNumber), Nationality: strings.TrimSpace(request.Nationality)}
	if err := h.db.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			respond.Error(w, r, 409, "EMAIL_EXISTS", "Ya existe una cuenta con ese correo", nil)
			return
		}
		respond.Error(w, r, 500, "INTERNAL_ERROR", "No se pudo crear el usuario", nil)
		return
	}
	respond.JSON(w, http.StatusCreated, user.Public(), "Usuario creado")
}
