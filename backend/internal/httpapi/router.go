package httpapi

import (
	"log/slog"
	"net/http"

	"starter-backend/internal/config"
	"starter-backend/internal/handlers"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg config.Config, logger *slog.Logger) (http.Handler, error) {
	router := mux.NewRouter()
	authenticator := middleware.NewAuthenticator(db, cfg)
	authHandler, err := handlers.NewAuthHandler(db, cfg)
	if err != nil {
		return nil, err
	}
	profileHandler := handlers.NewProfileHandler(db)
	adminHandler := handlers.NewAdminHandler(db)
	agencyHandler := handlers.NewAgencyHandler(db)

	router.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.NewRateLimiter(cfg.APIRatePerMinute, cfg.TrustProxy).Middleware)

	publicAuth := api.PathPrefix("/auth").Subrouter()
	publicAuth.Use(middleware.NewRateLimiter(cfg.AuthRatePerMinute, cfg.TrustProxy).Middleware)
	publicAuth.Use(authenticator.OriginGuard)
	publicAuth.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
	publicAuth.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
	publicAuth.HandleFunc("/options", authHandler.Options).Methods(http.MethodGet)
	publicAuth.HandleFunc("/google/start", authHandler.GoogleStart).Methods(http.MethodPost)
	// Google's top-level redirect carries no application Origin; OAuth state binds the browser.
	publicAuth.HandleFunc("/google/callback", authHandler.GoogleCallback).Methods(http.MethodGet)

	protected := api.NewRoute().Subrouter()
	protected.Use(authenticator.Middleware)
	protected.HandleFunc("/auth/session", authHandler.Session).Methods(http.MethodGet)
	protected.HandleFunc("/auth/logout", authHandler.Logout).Methods(http.MethodPost)
	protected.HandleFunc("/me", profileHandler.Get).Methods(http.MethodGet)
	protected.HandleFunc("/me", profileHandler.Update).Methods(http.MethodPatch)
	protected.HandleFunc("/me/password", authHandler.ChangePassword).Methods(http.MethodPost)
	protected.HandleFunc("/me/google/start", authHandler.GoogleLink).Methods(http.MethodPost)
	agency := protected.PathPrefix("/agency").Subrouter()
	agency.Use(middleware.RequireRoles(models.RoleAgency))
	agency.HandleFunc("", agencyHandler.Mine).Methods(http.MethodGet)
	agency.HandleFunc("", agencyHandler.UpdateMine).Methods(http.MethodPut)
	agency.HandleFunc("/options", agencyHandler.Options).Methods(http.MethodGet)

	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireRoles(models.RoleAdmin))
	admin.HandleFunc("/dashboard", adminHandler.Dashboard).Methods(http.MethodGet)
	admin.HandleFunc("/agencies/options", agencyHandler.Options).Methods(http.MethodGet)
	admin.HandleFunc("/agencies/managers", agencyHandler.Managers).Methods(http.MethodGet)
	admin.HandleFunc("/agencies", agencyHandler.List).Methods(http.MethodGet)
	admin.HandleFunc("/agencies", agencyHandler.Create).Methods(http.MethodPost)
	admin.HandleFunc("/agencies/{id:[0-9]+}", agencyHandler.Get).Methods(http.MethodGet)
	admin.HandleFunc("/agencies/{id:[0-9]+}", agencyHandler.Update).Methods(http.MethodPut)
	admin.HandleFunc("/users", adminHandler.ListUsers).Methods(http.MethodGet)
	admin.HandleFunc("/users", authHandler.CreateUser).Methods(http.MethodPost)
	admin.HandleFunc("/users/{id:[0-9]+}", adminHandler.UpdateUser).Methods(http.MethodPatch)
	admin.HandleFunc("/users/{id:[0-9]+}", adminHandler.GetUser).Methods(http.MethodGet)
	admin.HandleFunc("/users/{id:[0-9]+}/role", adminHandler.UpdateRole).Methods(http.MethodPatch)
	admin.HandleFunc("/users/{id:[0-9]+}/status", adminHandler.UpdateStatus).Methods(http.MethodPatch)

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", "Route was not found", nil)
	})
	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond.Error(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "HTTP method is not allowed", nil)
	})

	var handler http.Handler = router
	handler = middleware.CORS(cfg.AllowedOrigins).Handler(handler)
	handler = middleware.AccessLog(logger, cfg.TrustProxy)(handler)
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.Recover(logger)(handler)
	handler = middleware.RequestID(handler)
	return handler, nil
}
