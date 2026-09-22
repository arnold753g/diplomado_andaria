package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"starter-backend/internal/config"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	"starter-backend/internal/security"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthHandler struct {
	db           *gorm.DB
	cfg          config.Config
	validate     *validator.Validate
	dummyHash    string
	googleClient *http.Client
}

func NewAuthHandler(db *gorm.DB, cfg config.Config) (*AuthHandler, error) {
	dummyHash, err := security.HashPassword("not a valid account password", cfg.PasswordBcryptCost)
	if err != nil {
		return nil, err
	}
	return &AuthHandler{db: db, cfg: cfg, validate: validator.New(), dummyHash: dummyHash}, nil
}

type registerRequest struct {
	Email          string `json:"email" validate:"required,email,max=320"`
	Password       string `json:"password" validate:"required"`
	FirstName      string `json:"first_name" validate:"required,min=2,max=100"`
	LastName       string `json:"last_name" validate:"required,min=2,max=100"`
	Phone          string `json:"phone" validate:"omitempty,max=30"`
	DocumentNumber string `json:"document_number" validate:"omitempty,max=40"`
	Nationality    string `json:"nationality" validate:"omitempty,max=80"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email,max=320"`
	Password string `json:"password" validate:"required,max=128"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"max=128"`
	NewPassword     string `json:"new_password" validate:"required,max=128"`
}

type sessionResponse struct {
	User               models.PublicUser `json:"user"`
	CSRFToken          string            `json:"csrf_token"`
	SessionExpiresAt   time.Time         `json:"session_expires_at"`
	IdleTimeoutSeconds int64             `json:"idle_timeout_seconds"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.RegistrationEnabled {
		respond.Error(w, r, http.StatusForbidden, "REGISTRATION_DISABLED", "Public registration is disabled", nil)
		return
	}
	var request registerRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Email = normalizeEmail(request.Email)
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return
	}
	hash, err := security.HashPassword(request.Password, h.cfg.PasswordBcryptCost)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PASSWORD_POLICY", security.ErrPasswordPolicy.Error(), map[string]string{"password": security.ErrPasswordPolicy.Error()})
		return
	}
	user := models.User{Email: request.Email, PasswordHash: hash, FirstName: request.FirstName,
		LastName: request.LastName, Role: models.RoleUser, Status: models.StatusActive,
		Phone: strings.TrimSpace(request.Phone), DocumentNumber: strings.TrimSpace(request.DocumentNumber), Nationality: strings.TrimSpace(request.Nationality)}
	if err := h.db.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			respond.Error(w, r, http.StatusConflict, "EMAIL_EXISTS", "An account with that email already exists", nil)
			return
		}
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "The account could not be created", nil)
		return
	}
	respond.JSON(w, http.StatusCreated, user.Public(), "Account created")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Email = normalizeEmail(request.Email)
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return
	}

	var user models.User
	err := h.db.Where("email = ?", request.Email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_ = security.CheckPassword(request.Password, h.dummyHash)
		respond.Error(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect", nil)
		return
	}
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Login could not be completed", nil)
		return
	}
	now := time.Now().UTC()
	passwordValid := security.CheckPassword(request.Password, user.PasswordHash)
	if user.LockedUntil != nil && now.Before(*user.LockedUntil) && passwordValid {
		respond.Error(w, r, http.StatusTooManyRequests, "ACCOUNT_LOCKED", "Too many failed attempts; try again later", nil)
		return
	}
	if !passwordValid {
		attempts := user.FailedLoginAttempts + 1
		updates := map[string]any{"failed_login_attempts": attempts}
		if attempts >= h.cfg.LoginMaxAttempts {
			updates["locked_until"] = now.Add(h.cfg.LoginLockout)
			updates["failed_login_attempts"] = 0
		}
		_ = h.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error
		respond.Error(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect", nil)
		return
	}
	if user.Status != models.StatusActive {
		respond.Error(w, r, http.StatusForbidden, "ACCOUNT_INACTIVE", "This account is inactive", nil)
		return
	}
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	user.LastLoginAt = &now
	if err := h.db.Model(&user).Updates(map[string]any{"failed_login_attempts": 0, "locked_until": nil, "last_login_at": now}).Error; err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Login could not be completed", nil)
		return
	}
	response, err := h.issueSession(w, r, &user)
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Session could not be created", nil)
		return
	}
	respond.JSON(w, http.StatusOK, response, "Signed in")
}

func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return
	}
	var session models.Session
	if err := h.db.Where("id = ? AND revoked_at IS NULL", principal.SessionID).First(&session).Error; err != nil {
		respond.Error(w, r, http.StatusUnauthorized, "INVALID_SESSION", "Session is invalid or expired", nil)
		return
	}
	cookie, err := r.Cookie(security.CookieName)
	if err != nil {
		respond.Error(w, r, http.StatusUnauthorized, "INVALID_SESSION", "Session is invalid or expired", nil)
		return
	}
	csrf, err := security.CSRFToken(h.cfg.SessionSecret, cookie.Value)
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Session could not be restored", nil)
		return
	}
	respond.JSON(w, http.StatusOK, sessionResponse{User: principal.User.Public(), CSRFToken: csrf,
		SessionExpiresAt: session.ExpiresAt, IdleTimeoutSeconds: int64(h.cfg.SessionIdleTimeout.Seconds())}, "Session active")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.CurrentPrincipal(r)
	now := time.Now().UTC()
	_ = h.db.Model(&models.Session{}).Where("id = ? AND revoked_at IS NULL", principal.SessionID).Update("revoked_at", now).Error
	h.clearCookie(w)
	respond.JSON(w, http.StatusOK, nil, "Signed out")
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return
	}
	var request changePasswordRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return
	}
	if (principal.User.PasswordHash != "" && !security.CheckPassword(request.CurrentPassword, principal.User.PasswordHash)) || (principal.User.PasswordHash == "" && principal.User.GoogleSubject == nil) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "CURRENT_PASSWORD_INVALID", "La contraseña actual no es correcta", nil)
		return
	}
	if request.CurrentPassword == request.NewPassword {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PASSWORD_REUSED", "New password must be different", map[string]string{"newpassword": "Use a different password"})
		return
	}
	hash, err := security.HashPassword(request.NewPassword, h.cfg.PasswordBcryptCost)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PASSWORD_POLICY", security.ErrPasswordPolicy.Error(), map[string]string{"newpassword": security.ErrPasswordPolicy.Error()})
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, principal.User.ID).Error; err != nil {
			return err
		}
		if current.PasswordHash != principal.User.PasswordHash || current.Status != models.StatusActive {
			return errors.New("account changed during password update")
		}
		if err := tx.Model(&models.User{}).Where("id = ?", principal.User.ID).Update("password_hash", hash).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		return tx.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", principal.User.ID).Update("revoked_at", now).Error
	}); err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Password could not be changed", nil)
		return
	}
	h.clearCookie(w)
	respond.JSON(w, http.StatusOK, nil, "Password changed; sign in again")
}

func (h *AuthHandler) issueSession(w http.ResponseWriter, r *http.Request, user *models.User) (sessionResponse, error) {
	sessionID, err := security.RandomToken()
	if err != nil {
		return sessionResponse{}, err
	}
	rawToken, err := security.RandomToken()
	if err != nil {
		return sessionResponse{}, err
	}
	now := time.Now().UTC()
	expiresAt := now.Add(h.cfg.SessionTTL)
	userAgent := r.UserAgent()
	if len(userAgent) > 512 {
		userAgent = userAgent[:512]
	}
	session := models.Session{ID: sessionID, UserID: user.ID, TokenHash: security.HashToken(rawToken),
		LastActivityAt: now, ExpiresAt: expiresAt, UserAgent: userAgent,
		IPAddress: middleware.ClientIP(r, h.cfg.TrustProxy)}
	// Serialize issuing a session with password/role/status changes. A login that
	// authenticated old account state must not create a session after revocation.
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, user.ID).Error; err != nil {
			return err
		}
		if current.Status != models.StatusActive || current.PasswordHash != user.PasswordHash || current.Role != user.Role {
			return errors.New("account changed during login")
		}
		*user = current
		return tx.Create(&session).Error
	}); err != nil {
		return sessionResponse{}, err
	}
	csrf, err := security.CSRFToken(h.cfg.SessionSecret, rawToken)
	if err != nil {
		return sessionResponse{}, err
	}
	h.setCookie(w, rawToken, expiresAt)
	return sessionResponse{User: user.Public(), CSRFToken: csrf, SessionExpiresAt: expiresAt,
		IdleTimeoutSeconds: int64(h.cfg.SessionIdleTimeout.Seconds())}, nil
}

func (h *AuthHandler) setCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}
	http.SetCookie(w, &http.Cookie{Name: security.CookieName, Value: token, Path: "/",
		Expires: expiresAt, MaxAge: maxAge, HttpOnly: true, Secure: h.cfg.CookieSecure, SameSite: h.cfg.CookieSameSite})
}

func (h *AuthHandler) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: security.CookieName, Value: "", Path: "/",
		Expires: time.Unix(1, 0), MaxAge: -1, HttpOnly: true, Secure: h.cfg.CookieSecure, SameSite: h.cfg.CookieSameSite})
}
