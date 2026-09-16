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
	attractionHandler := handlers.NewAttractionHandler(db)
	packageHandler := handlers.NewPackageHandler(db)
	purchaseHandler := handlers.NewPurchaseHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)

	router.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.NewRateLimiter(cfg.APIRatePerMinute, cfg.TrustProxy).Middleware)

	api.HandleFunc("/attractions/options", attractionHandler.Options).Methods(http.MethodGet)
	api.HandleFunc("/attractions", attractionHandler.ListPublic).Methods(http.MethodGet)
	api.HandleFunc("/attractions/{id:[0-9]+}", attractionHandler.GetPublic).Methods(http.MethodGet)
	api.HandleFunc("/attractions/{id:[0-9]+}/photos/{photo:[0-9]+}", attractionHandler.PhotoPublic).Methods(http.MethodGet)
	api.HandleFunc("/packages/options", packageHandler.PublicOptions).Methods(http.MethodGet)
	api.HandleFunc("/packages", packageHandler.ListPublic).Methods(http.MethodGet)
	api.HandleFunc("/packages/{id:[0-9]+}", packageHandler.GetPublic).Methods(http.MethodGet)
	api.HandleFunc("/packages/{id:[0-9]+}/photos/{photo:[0-9]+}", packageHandler.PhotoPublic).Methods(http.MethodGet)

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
	accountDashboard := protected.PathPrefix("/me/dashboard").Subrouter()
	accountDashboard.Use(middleware.RequireRoles(models.RoleUser, models.RoleAgency, models.RoleAttraction))
	accountDashboard.HandleFunc("", dashboardHandler.Mine).Methods(http.MethodGet)
	agency := protected.PathPrefix("/agency").Subrouter()
	agency.Use(middleware.RequireRoles(models.RoleAgency))
	agency.HandleFunc("", agencyHandler.Mine).Methods(http.MethodGet)
	agency.HandleFunc("", agencyHandler.UpdateMine).Methods(http.MethodPut)
	agency.HandleFunc("/options", agencyHandler.Options).Methods(http.MethodGet)
	agency.HandleFunc("/packages/attractions", packageHandler.Attractions).Methods(http.MethodGet)
	agency.HandleFunc("/packages", packageHandler.ListMine).Methods(http.MethodGet)
	agency.HandleFunc("/packages", packageHandler.Create).Methods(http.MethodPost)
	agency.HandleFunc("/packages/{id:[0-9]+}", packageHandler.GetMine).Methods(http.MethodGet)
	agency.HandleFunc("/packages/{id:[0-9]+}", packageHandler.Update).Methods(http.MethodPut)
	agency.HandleFunc("/packages/{id:[0-9]+}/photos/{photo:[0-9]+}", packageHandler.PhotoMine).Methods(http.MethodGet)
	agency.HandleFunc("/packages/{id:[0-9]+}/departures/{departure:[0-9]+}", packageHandler.UpdateDeparture).Methods(http.MethodPatch)
	agency.HandleFunc("/packages/{id:[0-9]+}/departures/{departure:[0-9]+}/cancel", packageHandler.CancelDeparture).Methods(http.MethodPost)
	agency.HandleFunc("/packages/{id:[0-9]+}/departures/{departure:[0-9]+}/minimum-refund", packageHandler.ConfirmMinimumRefund).Methods(http.MethodPost)
	agency.HandleFunc("/purchases", purchaseHandler.ListAgency).Methods(http.MethodGet)
	agency.HandleFunc("/purchases/{id:[0-9]+}", purchaseHandler.GetAgency).Methods(http.MethodGet)
	agency.HandleFunc("/purchases/{id:[0-9]+}/proof", purchaseHandler.ProofAgency).Methods(http.MethodGet)
	agency.HandleFunc("/purchases/{id:[0-9]+}/review", purchaseHandler.Review).Methods(http.MethodPatch)
	agency.HandleFunc("/purchases/{id:[0-9]+}/refund-qr", purchaseHandler.RefundQRAgency).Methods(http.MethodGet)
	agency.HandleFunc("/purchases/{id:[0-9]+}/refund-proof", purchaseHandler.RefundProofAgency).Methods(http.MethodGet)
	agency.HandleFunc("/purchases/{id:[0-9]+}/refund", purchaseHandler.CompleteRefund).Methods(http.MethodPatch)

	managed := protected.PathPrefix("/managed-attractions").Subrouter()
	managed.Use(middleware.RequireRoles(models.RoleAttraction))
	managed.HandleFunc("", attractionHandler.ListMine).Methods(http.MethodGet)
	managed.HandleFunc("/{id:[0-9]+}", attractionHandler.GetMine).Methods(http.MethodGet)
	managed.HandleFunc("/{id:[0-9]+}", attractionHandler.UpdateMine).Methods(http.MethodPut)
	managed.HandleFunc("/{id:[0-9]+}/photos/{photo:[0-9]+}", attractionHandler.PhotoMine).Methods(http.MethodGet)
	favorites := protected.PathPrefix("/me/favorites").Subrouter()
	favorites.Use(middleware.RequireRoles(models.RoleUser))
	favorites.HandleFunc("", attractionHandler.Favorites).Methods(http.MethodGet)
	favorites.HandleFunc("/{id:[0-9]+}", attractionHandler.Favorite).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	touristPackages := protected.PathPrefix("/me/packages").Subrouter()
	touristPackages.Use(middleware.RequireRoles(models.RoleUser))
	touristPackages.HandleFunc("/{id:[0-9]+}/payment-options", purchaseHandler.PaymentOptions).Methods(http.MethodGet)
	touristPurchases := protected.PathPrefix("/me/purchases").Subrouter()
	touristPurchases.Use(middleware.RequireRoles(models.RoleUser))
	touristPurchases.HandleFunc("", purchaseHandler.ListMine).Methods(http.MethodGet)
	touristPurchases.HandleFunc("", purchaseHandler.Create).Methods(http.MethodPost)
	touristPurchases.HandleFunc("/{id:[0-9]+}", purchaseHandler.GetMine).Methods(http.MethodGet)
	touristPurchases.HandleFunc("/{id:[0-9]+}/proof", purchaseHandler.ProofMine).Methods(http.MethodGet)
	touristPurchases.HandleFunc("/{id:[0-9]+}/proof", purchaseHandler.UpdateProof).Methods(http.MethodPatch)
	touristPurchases.HandleFunc("/{id:[0-9]+}/cancel", purchaseHandler.CancelMine).Methods(http.MethodPost)
	touristPurchases.HandleFunc("/{id:[0-9]+}/refund-destination", purchaseHandler.UpdateRefundDestination).Methods(http.MethodPatch)
	touristPurchases.HandleFunc("/{id:[0-9]+}/refund-qr", purchaseHandler.RefundQRMine).Methods(http.MethodGet)
	touristPurchases.HandleFunc("/{id:[0-9]+}/refund-proof", purchaseHandler.RefundProofMine).Methods(http.MethodGet)

	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireRoles(models.RoleAdmin))
	admin.HandleFunc("/attractions/managers", attractionHandler.Managers).Methods(http.MethodGet)
	admin.HandleFunc("/attractions", attractionHandler.ListAdmin).Methods(http.MethodGet)
	admin.HandleFunc("/attractions", attractionHandler.Create).Methods(http.MethodPost)
	admin.HandleFunc("/attractions/{id:[0-9]+}", attractionHandler.GetAdmin).Methods(http.MethodGet)
	admin.HandleFunc("/attractions/{id:[0-9]+}", attractionHandler.UpdateAdmin).Methods(http.MethodPut)
	admin.HandleFunc("/attractions/{id:[0-9]+}/photos/{photo:[0-9]+}", attractionHandler.PhotoAdmin).Methods(http.MethodGet)
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
