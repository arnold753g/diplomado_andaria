package handlers

import (
	"cloud.google.com/go/auth/credentials/idtoken"
	"context"
	"errors"
	"net/http"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	"strings"
	"time"
)

func (h *AuthHandler) MobileOptions(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]any{
		"registration_enabled": h.cfg.RegistrationEnabled,
		"google_enabled":       h.cfg.GoogleClientID != "",
		// OAuth client identifiers are public; the client secret stays on the server.
		"google_server_client_id": h.cfg.GoogleClientID,
	}, "")
}

func googleTokenIdentity(ctx context.Context, token, audience string) (googleIdentity, error) {
	payload, err := idtoken.Validate(ctx, token, audience)
	if err != nil {
		return googleIdentity{}, err
	}
	return googlePayloadIdentity(payload, audience)
}

func googlePayloadIdentity(payload *idtoken.Payload, audience string) (googleIdentity, error) {
	if payload == nil || audience == "" || payload.Audience != audience ||
		(payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com") ||
		payload.Expires <= time.Now().Unix() || payload.Subject == "" || len(payload.Subject) > 255 {
		return googleIdentity{}, errors.New("invalid Google identity claims")
	}
	verified, _ := payload.Claims["email_verified"].(bool)
	email, _ := payload.Claims["email"].(string)
	email = normalizeEmail(email)
	if !verified || email == "" || len(email) > 320 {
		return googleIdentity{}, errors.New("unverified Google email")
	}
	given, _ := payload.Claims["given_name"].(string)
	family, _ := payload.Claims["family_name"].(string)
	return googleIdentity{Subject: payload.Subject, Email: email, EmailVerified: true, GivenName: given, FamilyName: family}, nil
}

// Android sends the Google-signed ID token. It never sends trusted profile data.
// Andaria then issues its existing revocable cookie session, as for web login.
func (h *AuthHandler) GoogleMobileLogin(w http.ResponseWriter, r *http.Request) {
	if h.cfg.GoogleClientID == "" {
		respond.Error(w, r, http.StatusServiceUnavailable, "GOOGLE_DISABLED", "El acceso con Google todavía no está configurado", nil)
		return
	}
	var input struct {
		IDToken string `json:"id_token"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if strings.TrimSpace(input.IDToken) == "" || len(input.IDToken) > 16384 {
		respond.Error(w, r, http.StatusUnprocessableEntity, "GOOGLE_TOKEN_INVALID", "No se pudo validar el acceso con Google", nil)
		return
	}
	verify := h.googleMobileVerifier
	if verify == nil {
		verify = googleTokenIdentity
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	identity, err := verify(ctx, input.IDToken, h.cfg.GoogleClientID)
	if err != nil || !identity.EmailVerified || h.validate.Var(identity.Email, "required,email,max=320") != nil {
		respond.Error(w, r, http.StatusUnauthorized, "GOOGLE_TOKEN_INVALID", "No se pudo validar el acceso con Google", nil)
		return
	}
	user, err := h.resolveGoogleUser(identity, oauthAttempt{})
	if err != nil {
		switch err.Error() {
		case "email_exists":
			respond.Error(w, r, http.StatusConflict, "GOOGLE_EMAIL_EXISTS", "Este correo ya tiene una cuenta. Ingresa con contraseña y vincula Google desde la web", nil)
		case "registration_disabled":
			respond.Error(w, r, http.StatusForbidden, "REGISTRATION_DISABLED", "El registro no está disponible", nil)
		case "inactive":
			respond.Error(w, r, http.StatusForbidden, "ACCOUNT_INACTIVE", "Esta cuenta está desactivada", nil)
		default:
			respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "No se pudo iniciar sesión", nil)
		}
		return
	}
	if user.Role != models.RoleUser {
		respond.Error(w, r, http.StatusForbidden, "TOURIST_REQUIRED", "Esta app es exclusiva para turistas. Usa la web para la gestión de otros roles", nil)
		return
	}
	session, err := h.issueSession(w, r, &user)
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "No se pudo iniciar sesión", nil)
		return
	}
	respond.JSON(w, http.StatusOK, session, "Sesión iniciada")
}
