package handlers

import (
	"context"
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
	db                   *gorm.DB
	cfg                  config.Config
	validate             *validator.Validate
	dummyHash            string
	googleClient         *http.Client
	googleMobileVerifier func(context.Context, string, string) (googleIdentity, error)
	sendVerification     func(context.Context, string, string) error
}

func NewAuthHandler(db *gorm.DB, cfg config.Config) (*AuthHandler, error) {
	dummyHash, err := security.HashPassword("not a valid account password", cfg.PasswordBcryptCost)
	if err != nil {
		return nil, err
	}
	h := &AuthHandler{db: db, cfg: cfg, validate: validator.New(), dummyHash: dummyHash}
	h.sendVerification = h.sendVerificationMail
	return h, nil
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
	if h.cfg.EmailVerificationRequired {
		h.registerPending(w, r, request, hash)
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
	err := h.db.Where("correo = ?", request.Email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_ = security.CheckPassword(request.Password, h.dummyHash)
		respond.Error(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect", nil)
		return
	}
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Login could not be completed", nil)
		return
	}
	passwordValid := security.CheckPassword(request.Password, user.PasswordHash)
	status := http.StatusUnauthorized
	// Hashing stays outside the row lock. Recheck the hash under the lock so a
	// concurrent password change cannot authenticate stale credentials.
	err = h.db.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, user.ID).Error; err != nil {
			return err
		}
		if current.PasswordHash != user.PasswordHash {
			return nil
		}
		now := time.Now().UTC()
		if current.LockedUntil != nil && now.Before(*current.LockedUntil) {
			if passwordValid {
				status = http.StatusTooManyRequests
			}
			return nil // Failed requests must not prolong an existing lockout.
		}
		if !passwordValid {
			attempts := current.FailedLoginAttempts + 1
			updates := map[string]any{"intentos_fallidos_acceso": attempts, "bloqueado_hasta": nil}
			if attempts >= h.cfg.LoginMaxAttempts {
				updates["bloqueado_hasta"] = now.Add(h.cfg.LoginLockout)
				updates["intentos_fallidos_acceso"] = 0
			}
			return tx.Model(&current).Updates(updates).Error
		}
		if current.Status != models.StatusActive {
			status = http.StatusForbidden
			return nil
		}
		user = current
		status = http.StatusOK
		return tx.Model(&current).Updates(map[string]any{"intentos_fallidos_acceso": 0, "bloqueado_hasta": nil, "ultimo_acceso_en": now}).Error
	})
	if err != nil {
		respond.Error(w, r, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Login could not be completed; try again later", nil)
		return
	}
	if status == http.StatusUnauthorized {
		respond.Error(w, r, status, "INVALID_CREDENTIALS", "Email or password is incorrect", nil)
		return
	}
	if status == http.StatusTooManyRequests {
		respond.Error(w, r, status, "ACCOUNT_LOCKED", "Too many failed attempts; try again later", nil)
		return
	}
	if status == http.StatusForbidden {
		respond.Error(w, r, http.StatusForbidden, "ACCOUNT_INACTIVE", "This account is inactive", nil)
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
	if err := h.db.Where("id = ? AND revocado_en IS NULL", principal.SessionID).First(&session).Error; err != nil {
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
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return
	}
	now := time.Now().UTC()
	err := h.db.WithContext(r.Context()).Model(&models.Session{}).Where("id = ? AND revocado_en IS NULL", principal.SessionID).Update("revocado_en", now).Error
	h.clearCookie(w)
	if err != nil {
		respond.Error(w, r, http.StatusServiceUnavailable, "LOGOUT_UNAVAILABLE", "Se borró el acceso local, pero no se pudo confirmar el cierre de sesión en el servidor", nil)
		return
	}
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
		if err := tx.Model(&models.User{}).Where("id = ?", principal.User.ID).Update("hash_contrasena", hash).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		return tx.Model(&models.Session{}).Where("usuario_id = ? AND revocado_en IS NULL", principal.User.ID).Update("revocado_en", now).Error
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
