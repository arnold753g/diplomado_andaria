package handlers

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	"starter-backend/internal/security"
)

const oauthCookie = "andaria_google_state"

type oauthAttempt struct {
	StateHash string `gorm:"primaryKey"`
	Verifier  string
	UserID    *uint64
	SessionID *string
	ExpiresAt time.Time
}

func (oauthAttempt) TableName() string { return "oauth_attempts" }

type googleIdentity struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func (h *AuthHandler) Options(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	respond.JSON(w, 200, map[string]bool{"registration_enabled": h.cfg.RegistrationEnabled, "google_enabled": h.cfg.GoogleClientID != ""}, "")
}
func (h *AuthHandler) GoogleStart(w http.ResponseWriter, r *http.Request) {
	h.startGoogle(w, r, nil, nil)
}
func (h *AuthHandler) GoogleLink(w http.ResponseWriter, r *http.Request) {
	p, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, 401, "UNAUTHORIZED", "Inicia sesión", nil)
		return
	}
	if p.User.GoogleSubject != nil {
		respond.Error(w, r, 409, "GOOGLE_ALREADY_LINKED", "La cuenta ya está vinculada", nil)
		return
	}
	h.startGoogle(w, r, &p.User.ID, &p.SessionID)
}
func (h *AuthHandler) startGoogle(w http.ResponseWriter, r *http.Request, userID *uint64, sessionID *string) {
	if h.cfg.GoogleClientID == "" {
		respond.Error(w, r, 503, "GOOGLE_DISABLED", "El acceso con Google todavía no está configurado", nil)
		return
	}
	state, err := security.RandomToken()
	if err != nil {
		respond.Error(w, r, 500, "INTERNAL_ERROR", "No se pudo iniciar Google", nil)
		return
	}
	verifier, err := security.RandomToken()
	if err != nil {
		respond.Error(w, r, 500, "INTERNAL_ERROR", "No se pudo iniciar Google", nil)
		return
	}
	attempt := oauthAttempt{StateHash: security.HashToken(state), Verifier: verifier, UserID: userID, SessionID: sessionID, ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("expires_at < ?", time.Now().UTC()).Delete(&oauthAttempt{}).Error; err != nil {
			return err
		}
		return tx.Create(&attempt).Error
	}); err != nil {
		respond.Error(w, r, 500, "INTERNAL_ERROR", "No se pudo iniciar Google", nil)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: oauthCookie, Value: state, Path: "/api/v1/auth/google", HttpOnly: true, Secure: h.cfg.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	challenge := sha256.Sum256([]byte(verifier))
	query := url.Values{"client_id": {h.cfg.GoogleClientID}, "redirect_uri": {h.cfg.GoogleRedirectURL}, "response_type": {"code"}, "scope": {"openid email profile"}, "state": {state}, "code_challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])}, "code_challenge_method": {"S256"}, "prompt": {"select_account"}}
	w.Header().Set("Cache-Control", "no-store")
	respond.JSON(w, 200, map[string]string{"url": "https://accounts.google.com/o/oauth2/v2/auth?" + query.Encode()}, "")
}

func (h *AuthHandler) googleRedirect(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, h.cfg.FrontendURL+"/auth/google?result="+url.QueryEscape(code), http.StatusSeeOther)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if h.cfg.GoogleClientID == "" {
		respond.Error(w, r, 503, "GOOGLE_DISABLED", "Google no está configurado", nil)
		return
	}
	cookie, err := r.Cookie(oauthCookie)
	state := r.URL.Query().Get("state")
	if err != nil || len(state) != 64 || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		h.googleRedirect(w, r, "invalid_state")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: oauthCookie, Path: "/api/v1/auth/google", MaxAge: -1, HttpOnly: true, Secure: h.cfg.CookieSecure, SameSite: http.SameSiteLaxMode})
	var attempt oauthAttempt
	// Consumption is atomic: retries and concurrent callbacks cannot reuse an authorization attempt.
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("state_hash = ? AND expires_at > ?", security.HashToken(state), time.Now().UTC()).First(&attempt).Error; err != nil {
			return err
		}
		return tx.Delete(&attempt).Error
	})
	if err != nil {
		h.googleRedirect(w, r, "invalid_state")
		return
	}
	if r.URL.Query().Get("error") != "" {
		h.googleRedirect(w, r, "cancelled")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" || len(code) > 4096 {
		h.googleRedirect(w, r, "failed")
		return
	}
	identity, err := h.exchangeGoogle(r.Context(), code, attempt.Verifier)
	if err != nil {
		h.googleRedirect(w, r, "failed")
		return
	}
	user, err := h.resolveGoogleUser(identity, attempt)
	if err != nil {
		result := "failed"
		for _, known := range []string{"email_exists", "registration_disabled", "inactive", "link_invalid"} {
			if err.Error() == known {
				result = known
			}
		}
		h.googleRedirect(w, r, result)
		return
	}
	if attempt.UserID != nil {
		h.googleRedirect(w, r, "linked")
		return
	}
	if _, err := h.issueSession(w, r, &user); err != nil {
		h.googleRedirect(w, r, "failed")
		return
	}
	h.googleRedirect(w, r, "success")
}

// The code is exchanged server-to-server using client credentials and PKCE.
// Identity comes from Google's authenticated UserInfo endpoint, never from client claims.
func (h *AuthHandler) exchangeGoogle(ctx context.Context, code, verifier string) (googleIdentity, error) {
	client := h.googleClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	values := url.Values{"code": {code}, "client_id": {h.cfg.GoogleClientID}, "client_secret": {h.cfg.GoogleClientSecret}, "redirect_uri": {h.cfg.GoogleRedirectURL}, "grant_type": {"authorization_code"}, "code_verifier": {verifier}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(values.Encode()))
	if err != nil {
		return googleIdentity{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return googleIdentity{}, err
	}
	defer res.Body.Close()
	var token struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if res.StatusCode != 200 {
		return googleIdentity{}, errors.New("token exchange failed")
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&token); err != nil {
		return googleIdentity{}, err
	}
	if token.AccessToken == "" || !strings.EqualFold(token.TokenType, "Bearer") {
		return googleIdentity{}, errors.New("invalid token response")
	}
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return googleIdentity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	res, err = client.Do(req)
	if err != nil {
		return googleIdentity{}, err
	}
	defer res.Body.Close()
	var identity googleIdentity
	if res.StatusCode != 200 {
		return identity, errors.New("userinfo failed")
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&identity); err != nil {
		return identity, err
	}
	identity.Email = normalizeEmail(identity.Email)
	if !identity.EmailVerified || identity.Subject == "" || len(identity.Subject) > 255 || h.validate.Var(identity.Email, "required,email,max=320") != nil {
		return identity, errors.New("unverified identity")
	}
	return identity, nil
}

func (h *AuthHandler) resolveGoogleUser(identity googleIdentity, attempt oauthAttempt) (models.User, error) {
	var user models.User
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// Serializes identity creation/linking, including same-email races between Google callbacks.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(7412002)").Error; err != nil {
			return err
		}
		if attempt.UserID != nil {
			if attempt.SessionID == nil {
				return errors.New("link_invalid")
			}
			var session models.Session
			if err := tx.Where("id = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ? AND last_activity_at > ?", *attempt.SessionID, *attempt.UserID, time.Now().UTC(), time.Now().UTC().Add(-h.cfg.SessionIdleTimeout)).First(&session).Error; err != nil {
				return errors.New("link_invalid")
			}
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, *attempt.UserID).Error; err != nil {
				return err
			}
			if user.Status != models.StatusActive {
				return errors.New("inactive")
			}
			if user.Email != identity.Email || user.GoogleSubject != nil {
				return errors.New("link_invalid")
			}
			if err := tx.Model(&user).Update("google_subject", identity.Subject).Error; err != nil {
				return errors.New("link_invalid")
			}
			return nil
		}
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("google_subject = ?", identity.Subject).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var count int64
			if err := tx.Model(&models.User{}).Where("email = ?", identity.Email).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New("email_exists")
			}
			if !h.cfg.RegistrationEnabled {
				return errors.New("registration_disabled")
			}
			user = models.User{Email: identity.Email, FirstName: googleName(identity.GivenName, "Viajero"), LastName: googleName(identity.FamilyName, ""), GoogleSubject: &identity.Subject, Role: models.RoleUser, Status: models.StatusActive}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if user.Status != models.StatusActive {
			return errors.New("inactive")
		}
		now := time.Now().UTC()
		user.LastLoginAt = &now
		return tx.Model(&user).Update("last_login_at", now).Error
	})
	return user, err
}
func googleName(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if utf8.RuneCountInString(value) > 100 {
		return string([]rune(value)[:100])
	}
	return value
}
