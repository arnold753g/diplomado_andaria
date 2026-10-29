package handlers

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	"starter-backend/internal/security"
)

type pendingRegistration struct {
	Email          string    `gorm:"column:correo;primaryKey"`
	TokenHash      string    `gorm:"column:hash_token"`
	PasswordHash   string    `gorm:"column:hash_contrasena"`
	FirstName      string    `gorm:"column:nombres"`
	LastName       string    `gorm:"column:apellidos"`
	Phone          string    `gorm:"column:telefono"`
	DocumentNumber string    `gorm:"column:numero_documento"`
	Nationality    string    `gorm:"column:nacionalidad"`
	CreatedAt      time.Time `gorm:"column:creado_en"`
	ExpiresAt      time.Time `gorm:"column:vence_en"`
}

func (pendingRegistration) TableName() string { return "registros_pendientes" }

var errRegistrationThrottled = errors.New("registration throttled")

func (h *AuthHandler) registerPending(w http.ResponseWriter, r *http.Request, request registerRequest, hash string) {
	if h.cfg.SMTPHost == "" {
		message := "El registro con contraseña no está disponible en este momento. Inténtalo más tarde."
		if h.cfg.GoogleClientID != "" {
			message += " También puedes continuar con Google."
		}
		respond.Error(w, r, 503, "EMAIL_UNAVAILABLE", message, nil)
		return
	}
	token, err := security.RandomToken()
	if err != nil {
		respond.Error(w, r, 503, "REGISTRATION_UNAVAILABLE", "No se pudo preparar el registro", nil)
		return
	}
	now := time.Now().UTC()
	pending := pendingRegistration{Email: request.Email, TokenHash: security.HashToken(token), PasswordHash: hash,
		FirstName: request.FirstName, LastName: request.LastName, Phone: strings.TrimSpace(request.Phone),
		DocumentNumber: strings.TrimSpace(request.DocumentNumber), Nationality: strings.TrimSpace(request.Nationality), CreatedAt: now, ExpiresAt: now.Add(30 * time.Minute)}
	deliver := false
	err = h.db.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		// Serialize requests even before an email has a row. A pending registration
		// never reserves a users.email or prevents the owner signing up via Google.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", request.Email).Error; err != nil {
			return err
		}
		var previous pendingRegistration
		err := tx.Where("correo = ?", request.Email).First(&previous).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && now.Sub(previous.CreatedAt) < time.Minute {
			return errRegistrationThrottled
		}
		var count int64
		if err := tx.Model(&models.User{}).Where("correo = ?", request.Email).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return nil
		} // Same response; do not enumerate existing users.
		if err := tx.Where("vence_en <= ?", now).Delete(&pendingRegistration{}).Error; err != nil {
			return err
		}
		deliver = true
		// GORM UpdateAll omits auto-create timestamps. Reset created_at explicitly
		// so every resend starts a fresh cooldown, not only the first request.
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "correo"}},
			DoUpdates: clause.AssignmentColumns([]string{"hash_token", "hash_contrasena", "nombres", "apellidos", "telefono", "numero_documento", "nacionalidad", "creado_en", "vence_en"}),
		}).Create(&pending).Error
	})
	if errors.Is(err, errRegistrationThrottled) {
		w.Header().Set("Retry-After", "60")
		respond.Error(w, r, 429, "REGISTRATION_THROTTLED", "Espera un minuto antes de solicitar otro enlace", nil)
		return
	}
	if err != nil {
		respond.Error(w, r, 503, "REGISTRATION_UNAVAILABLE", "No se pudo preparar el registro", nil)
		return
	}
	if deliver {
		link := strings.TrimRight(h.cfg.FrontendURL, "/") + "/verify-email#token=" + url.QueryEscape(token)
		if err := h.sendVerification(r.Context(), request.Email, link); err != nil {
			// Keep the failed row for the cooldown. Its token is never returned to clients.
			respond.Error(w, r, 503, "EMAIL_UNAVAILABLE", "No se pudo enviar el correo. Inténtalo nuevamente en un minuto", nil)
			return
		}
	}
	respond.JSON(w, http.StatusAccepted, map[string]bool{"verification_required": true}, "Si el correo permite el registro, recibirás un enlace para confirmar tu cuenta. Revisa también spam.")
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.RegistrationEnabled {
		respond.Error(w, r, http.StatusForbidden, "REGISTRATION_DISABLED", "Public registration is disabled", nil)
		return
	}
	var request struct {
		Token string `json:"token"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if len(request.Token) != 64 {
		respond.Error(w, r, 422, "INVALID_VERIFICATION", "El enlace no es válido o venció", nil)
		return
	}
	invalid := errors.New("invalid verification")
	err := h.db.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		var pending pendingRegistration
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("hash_token = ?", security.HashToken(request.Token)).First(&pending).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return invalid
		}
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if !now.Before(pending.ExpiresAt) {
			return invalid
		}
		user := models.User{Email: pending.Email, PasswordHash: pending.PasswordHash, FirstName: pending.FirstName, LastName: pending.LastName,
			Phone: pending.Phone, DocumentNumber: pending.DocumentNumber, Nationality: pending.Nationality, Role: models.RoleUser, Status: models.StatusActive, EmailVerifiedAt: &now}
		if err := tx.Create(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return invalid
			}
			return err
		}
		return tx.Where("hash_token = ?", pending.TokenHash).Delete(&pendingRegistration{}).Error
	})
	if errors.Is(err, invalid) {
		respond.Error(w, r, 422, "INVALID_VERIFICATION", "El enlace no es válido o venció. Si ya tienes cuenta, inicia sesión", nil)
		return
	}
	if err != nil {
		respond.Error(w, r, 503, "VERIFICATION_UNAVAILABLE", "No se pudo confirmar el correo. Inténtalo nuevamente", nil)
		return
	}
	respond.JSON(w, 200, map[string]bool{"verified": true}, "Correo confirmado. Ya puedes iniciar sesión en la web o la app.")
}

func (h *AuthHandler) sendVerificationMail(ctx context.Context, to, link string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	address := net.JoinHostPort(h.cfg.SMTPHost, h.cfg.SMTPPort)
	var conn net.Conn
	var err error
	tlsConfig := &tls.Config{ServerName: h.cfg.SMTPHost, MinVersion: tls.VersionTLS12}
	if h.cfg.SMTPPort == "465" {
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, "tcp", address)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, h.cfg.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()
	if h.cfg.SMTPPort != "465" {
		if err := client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}
	if h.cfg.SMTPUsername != "" {
		if err := client.Auth(smtp.PlainAuth("", h.cfg.SMTPUsername, h.cfg.SMTPPassword, h.cfg.SMTPHost)); err != nil {
			return err
		}
	}
	if err := client.Mail(h.cfg.SMTPFrom); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Confirma tu cuenta en Andaria\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nAbre este enlace para confirmar el registro. Vence en 30 minutos:\r\n%s\r\n\r\nSi no solicitaste esta cuenta, ignora este mensaje. No compartas el enlace.\r\n", h.cfg.SMTPFrom, to, link)
	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
