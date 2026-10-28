# Contrato API de Andaria

Revisión local del 11/10/2026 sobre commit base `2f26d105`, con cambios sin publicar. 81 combinaciones de método/ruta; 73 del alcance académico. Se excluyen del subconjunto de defensa las cuatro rutas de atractivos asignados y las cuatro combinaciones de favoritos.

Las tres rutas de salud ejecutan SELECT 1: 200 con `data.database=ok`; 503 ante fallo. La producción observada conserva el chequeo anterior `/health`; los alias devuelven 404 hasta redesplegar.

Sesión: cookie `starter_session` HttpOnly. Mutaciones autenticadas: `X-CSRF-Token`. Errores: 400 JSON/tipos inválidos, 422 validaciones, 401 sesión ausente/vencida, 403 rol o CSRF, 404 recurso ausente/ajeno, 409 estado/versión/cupos, 429 límite, 500 fallo interno sanitizado.

| Método | Ruta completa | Permiso | Handler |
|---|---|---|---|
| GET | /health | Público | NewHealth |
| GET | /api/v1/salud | Público | NewHealth |
| GET | /api/v1/health | Público | NewHealth |
| GET | /api/v1/attractions/options | Público | Options |
| GET | /api/v1/attractions | Público | ListPublic |
| GET | /api/v1/attractions/{id:[0-9]+} | Público | GetPublic |
| GET | /api/v1/attractions/{id:[0-9]+}/photos/{photo:[0-9]+} | Público | PhotoPublic |
| GET | /api/v1/packages/options | Público | PublicOptions |
| GET | /api/v1/packages | Público | ListPublic |
| GET | /api/v1/packages/{id:[0-9]+} | Público | GetPublic |
| GET | /api/v1/packages/{id:[0-9]+}/photos/{photo:[0-9]+} | Público | PhotoPublic |
| POST | /api/v1/auth/register | Público; guardia de origen | Register |
| POST | /api/v1/auth/login | Público; guardia de origen | Login |
| GET | /api/v1/auth/options | Público; guardia de origen | Options |
| GET | /api/v1/auth/mobile/options | Público; guardia de origen | MobileOptions |
| POST | /api/v1/auth/google/mobile | Público; guardia de origen | GoogleMobileLogin |
| POST | /api/v1/auth/google/start | Público; guardia de origen | GoogleStart |
| GET | /api/v1/auth/google/callback | Público; guardia de origen | GoogleCallback |
| GET | /api/v1/auth/session | Sesión válida | Session |
| POST | /api/v1/auth/logout | Sesión válida | Logout |
| GET | /api/v1/me | Sesión válida | Get |
| PATCH | /api/v1/me | Sesión válida | Update |
| POST | /api/v1/me/password | Sesión válida | ChangePassword |
| POST | /api/v1/me/google/start | Sesión válida | GoogleLink |
| GET | /api/v1/me/dashboard | Turista/agencia/atracción | Mine |
| GET | /api/v1/agency | Encargado de agencia; propia | Mine |
| PUT | /api/v1/agency | Encargado de agencia; propia | UpdateMine |
| GET | /api/v1/agency/options | Encargado de agencia; propia | Options |
| GET | /api/v1/agency/packages/attractions | Encargado de agencia; propia | Attractions |
| GET | /api/v1/agency/packages | Encargado de agencia; propia | ListMine |
| POST | /api/v1/agency/packages | Encargado de agencia; propia | Create |
| GET | /api/v1/agency/packages/{id:[0-9]+} | Encargado de agencia; propia | GetMine |
| PUT | /api/v1/agency/packages/{id:[0-9]+} | Encargado de agencia; propia | Update |
| GET | /api/v1/agency/packages/{id:[0-9]+}/photos/{photo:[0-9]+} | Encargado de agencia; propia | PhotoMine |
| PATCH | /api/v1/agency/packages/{id:[0-9]+}/departures/{departure:[0-9]+} | Encargado de agencia; propia | UpdateDeparture |
| POST | /api/v1/agency/packages/{id:[0-9]+}/departures/{departure:[0-9]+}/cancel | Encargado de agencia; propia | CancelDeparture |
| POST | /api/v1/agency/packages/{id:[0-9]+}/departures/{departure:[0-9]+}/minimum-refund | Encargado de agencia; propia | ConfirmMinimumRefund |
| GET | /api/v1/agency/purchases | Encargado de agencia; propia | ListAgency |
| GET | /api/v1/agency/purchases/{id:[0-9]+} | Encargado de agencia; propia | GetAgency |
| GET | /api/v1/agency/purchases/{id:[0-9]+}/proof | Encargado de agencia; propia | ProofAgency |
| PATCH | /api/v1/agency/purchases/{id:[0-9]+}/review | Encargado de agencia; propia | Review |
| GET | /api/v1/agency/purchases/{id:[0-9]+}/refund-qr | Encargado de agencia; propia | RefundQRAgency |
| GET | /api/v1/agency/purchases/{id:[0-9]+}/refund-proof | Encargado de agencia; propia | RefundProofAgency |
| PATCH | /api/v1/agency/purchases/{id:[0-9]+}/refund | Encargado de agencia; propia | CompleteRefund |
| GET | /api/v1/managed-attractions | Encargado de atracción; asignada | ListMine |
| GET | /api/v1/managed-attractions/{id:[0-9]+} | Encargado de atracción; asignada | GetMine |
| PUT | /api/v1/managed-attractions/{id:[0-9]+} | Encargado de atracción; asignada | UpdateMine |
| GET | /api/v1/managed-attractions/{id:[0-9]+}/photos/{photo:[0-9]+} | Encargado de atracción; asignada | PhotoMine |
| GET | /api/v1/me/favorites | Turista; propio | Favorites |
| GET | /api/v1/me/favorites/{id:[0-9]+} | Turista; propio | Favorite |
| PUT | /api/v1/me/favorites/{id:[0-9]+} | Turista; propio | Favorite |
| DELETE | /api/v1/me/favorites/{id:[0-9]+} | Turista; propio | Favorite |
| GET | /api/v1/me/packages/{id:[0-9]+}/payment-options | Turista | PaymentOptions |
| GET | /api/v1/me/purchases | Turista; propio | ListMine |
| POST | /api/v1/me/purchases | Turista; propio | Create |
| GET | /api/v1/me/purchases/{id:[0-9]+} | Turista; propio | GetMine |
| GET | /api/v1/me/purchases/{id:[0-9]+}/proof | Turista; propio | ProofMine |
| PATCH | /api/v1/me/purchases/{id:[0-9]+}/proof | Turista; propio | UpdateProof |
| POST | /api/v1/me/purchases/{id:[0-9]+}/cancel | Turista; propio | CancelMine |
| PATCH | /api/v1/me/purchases/{id:[0-9]+}/refund-destination | Turista; propio | UpdateRefundDestination |
| GET | /api/v1/me/purchases/{id:[0-9]+}/refund-qr | Turista; propio | RefundQRMine |
| GET | /api/v1/me/purchases/{id:[0-9]+}/refund-proof | Turista; propio | RefundProofMine |
| GET | /api/v1/admin/attractions/managers | Administrador | Managers |
| GET | /api/v1/admin/attractions | Administrador | ListAdmin |
| POST | /api/v1/admin/attractions | Administrador | Create |
| GET | /api/v1/admin/attractions/{id:[0-9]+} | Administrador | GetAdmin |
| PUT | /api/v1/admin/attractions/{id:[0-9]+} | Administrador | UpdateAdmin |
| GET | /api/v1/admin/attractions/{id:[0-9]+}/photos/{photo:[0-9]+} | Administrador | PhotoAdmin |
| GET | /api/v1/admin/dashboard | Administrador | Dashboard |
| GET | /api/v1/admin/agencies/options | Administrador | Options |
| GET | /api/v1/admin/agencies/managers | Administrador | Managers |
| GET | /api/v1/admin/agencies | Administrador | List |
| POST | /api/v1/admin/agencies | Administrador | Create |
| GET | /api/v1/admin/agencies/{id:[0-9]+} | Administrador | Get |
| PUT | /api/v1/admin/agencies/{id:[0-9]+} | Administrador | Update |
| GET | /api/v1/admin/users | Administrador | ListUsers |
| POST | /api/v1/admin/users | Administrador | CreateUser |
| PATCH | /api/v1/admin/users/{id:[0-9]+} | Administrador | UpdateUser |
| GET | /api/v1/admin/users/{id:[0-9]+} | Administrador | GetUser |
| PATCH | /api/v1/admin/users/{id:[0-9]+}/role | Administrador | UpdateRole |
| PATCH | /api/v1/admin/users/{id:[0-9]+}/status | Administrador | UpdateStatus |

## Contrato compacto

Las rutas siguientes usan base `/api/v1`. La colección importable está en `docs/api/Andaria.postman_collection.json`.

| Método | Ruta | Sesión / rol | Parámetros | Respuestas |
|---|---|---|---|---|
| POST | /auth/login | Pública | email, password (JSON) | 200; 400, 401, 403, 422, 429, 500 |
| GET | /attractions | Pública | page, limit, search, category | 200; 422, 429, 500 |
| GET | /packages | Pública | page, limit, search, department, bookable | 200; 422, 429, 500 |
| POST | /agency/packages | Agencia | name, duration_days, duration_nights; datos de oferta | 201; 400, 401, 403, 409, 422, 500 |
| POST | /me/purchases | Turista | departure_id, viajeros, medio y comprobante; clave opcional | 201; reintento 200; 400, 401, 403, 409, 422, 500 |
| PATCH | /agency/purchases/{id}/review | Agencia propia | id; version, decision, reason | 200; 400, 401, 403, 404, 409, 422, 500 |
| GET | /me/purchases/{id}/receipt | Turista propietario | id; download=1 opcional | PDF 200; 401, 403, 404, 409, 500 |
| GET | /agency/purchases/{id}/receipt | Agencia propia | id; download=1 opcional | PDF 200; 401, 403, 404, 409, 500 |
| POST | /me/purchases/{id}/cancel | Turista propio | id; version, reason, destino | 200; 400, 401, 403, 404, 409, 422, 500 |
| GET | /salud | Pública | Sin parámetros; SELECT 1 | 200; 503 |

## Estructuras de entrada y respuesta

Transcripción de los structs con campos JSON en handlers. `omitempty` indica omisión en serialización, no obligatoriedad de entrada; las restricciones adicionales se aplican en cada handler.

### agencyAdminRequest — agency.go

```go
type agencyAdminRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=160"`
	Description string  `json:"description" validate:"max=3000"`
	Department  string  `json:"department" validate:"required"`
	City        string  `json:"city" validate:"required,min=2,max=100"`
	Address     string  `json:"address" validate:"required,min=3,max=250"`
	Phone       string  `json:"phone" validate:"required,min=7,max=30"`
	Email       string  `json:"email" validate:"required,email,max=320"`
	ManagerID   *uint64 `json:"manager_id,omitempty"`
	Status      *string `json:"status,omitempty"`
	Published   bool    `json:"published"`
	Version     int     `json:"version"`
}
```

### agencyRequest — agency.go

```go
type agencyRequest struct {
	agencyAdminRequest
	MinimumPayingAge    *int   `json:"minimum_paying_age"`
	AcceptsQR           bool   `json:"accepts_qr"`
	AcceptsTransfer     bool   `json:"accepts_transfer"`
	BankName            string `json:"bank_name" validate:"max=100"`
	AccountHolder       string `json:"account_holder" validate:"max=160"`
	AccountNumber       string `json:"account_number" validate:"max=50"`
	PaymentInstructions string `json:"payment_instructions" validate:"max=1000"`
	QRImage             string `json:"qr_image"`
}
```

### agencyResponse — agency.go

```go
type agencyResponse struct {
	models.Agency
	QRImage       string `json:"qr_image,omitempty"`
	ManagerName   string `json:"manager_name"`
	ManagerEmail  string `json:"manager_email"`
	ManagerStatus string `json:"manager_status"`
}
```

### attractionInput — attraction.go

```go
type attractionInput struct {
	Name             string   `json:"name" validate:"required,min=2,max=160"`
	Description      string   `json:"description" validate:"max=6000"`
	SubcategoryIDs   []uint64 `json:"subcategory_ids"`
	ScheduleMode     string   `json:"schedule_mode"`
	OpeningTime      string   `json:"opening_time"`
	ClosingTime      string   `json:"closing_time"`
	OpeningDays      []int    `json:"opening_days"`
	SeasonMode       string   `json:"season_mode"`
	SeasonStartMonth *int     `json:"season_start_month"`
	SeasonEndMonth   *int     `json:"season_end_month"`
	Department       string   `json:"department"`
	City             string   `json:"city" validate:"required,min=2,max=100"`
	Address          string   `json:"address" validate:"required,min=3,max=250"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	AdmissionCents   int      `json:"admission_cents" validate:"min=0,max=100000000"`
	Recommendations  string   `json:"recommendations" validate:"max=3000"`
	Phone            string   `json:"phone" validate:"max=30"`
	ManagerID        *uint64  `json:"manager_id"`
	Status           *string  `json:"status"`
	Published        bool     `json:"published"`
	Version          int      `json:"version"`
	Photos           []struct {
		ID   uint64 `json:"id"`
		Data string `json:"data"`
	} `json:"photos"`
}
```

### attractionAdminView — attraction.go

```go
type attractionAdminView struct {
	models.Attraction
	ManagerID     uint64 `json:"manager_id"`
	ManagerName   string `json:"manager_name"`
	ManagerEmail  string `json:"manager_email"`
	ManagerStatus string `json:"manager_status"`
	Status        string `json:"status"`
	Published     bool   `json:"published"`
	Version       int    `json:"version"`
}
```

### registerRequest — auth.go

```go
type registerRequest struct {
	Email          string `json:"email" validate:"required,email,max=320"`
	Password       string `json:"password" validate:"required"`
	FirstName      string `json:"first_name" validate:"required,min=2,max=100"`
	LastName       string `json:"last_name" validate:"required,min=2,max=100"`
	Phone          string `json:"phone" validate:"omitempty,max=30"`
	DocumentNumber string `json:"document_number" validate:"omitempty,max=40"`
	Nationality    string `json:"nationality" validate:"omitempty,max=80"`
}
```

### loginRequest — auth.go

```go
type loginRequest struct {
	Email    string `json:"email" validate:"required,email,max=320"`
	Password string `json:"password" validate:"required,max=128"`
}
```

### changePasswordRequest — auth.go

```go
type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"max=128"`
	NewPassword     string `json:"new_password" validate:"required,max=128"`
}
```

### sessionResponse — auth.go

```go
type sessionResponse struct {
	User               models.PublicUser `json:"user"`
	CSRFToken          string            `json:"csrf_token"`
	SessionExpiresAt   time.Time         `json:"session_expires_at"`
	IdleTimeoutSeconds int64             `json:"idle_timeout_seconds"`
}
```

### dashboardNextPurchase — dashboard.go

```go
type dashboardNextPurchase struct {
	Reference      string    `json:"reference"`
	PackageName    string    `json:"package_name"`
	DepartureStart time.Time `json:"departure_start"`
	Status         string    `json:"status"`
}
```

### touristDashboard — dashboard.go

```go
type touristDashboard struct {
	Favorites        int64                  `json:"favorites"`
	ActivePurchases  int64                  `json:"active_purchases"`
	PendingRefunds   int64                  `json:"pending_refunds"`
	CompletedRefunds int64                  `json:"completed_refunds"`
	NextPurchase     *dashboardNextPurchase `json:"next_purchase,omitempty"`
}
```

### agencyDashboard — dashboard.go

```go
type agencyDashboard struct {
	Assigned          bool   `json:"assigned"`
	AgencyName        string `json:"agency_name,omitempty"`
	Packages          int64  `json:"packages"`
	PublishedPackages int64  `json:"published_packages"`
	PaymentsToReview  int64  `json:"payments_to_review"`
	MinimumReviews    int64  `json:"minimum_reviews"`
	PendingRefunds    int64  `json:"pending_refunds"`
	RefundsDueSoon    int64  `json:"refunds_due_soon"`
	OverdueRefunds    int64  `json:"overdue_refunds"`
}
```

### attractionManagerDashboard — dashboard.go

```go
type attractionManagerDashboard struct {
	AssignedAttractions  int64 `json:"assigned_attractions"`
	PublishedAttractions int64 `json:"published_attractions"`
	DraftAttractions     int64 `json:"draft_attractions"`
	InactiveAttractions  int64 `json:"inactive_attractions"`
}
```

### googleIdentity — google.go

```go
type googleIdentity struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}
```

### packagePhotoInput — package.go

```go
type packagePhotoInput struct {
	ID   uint64 `json:"id"`
	Data string `json:"data"`
}
```

### packageDayInput — package.go

```go
type packageDayInput struct {
	DayNumber     int      `json:"day_number"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Activities    []string `json:"activities"`
	AttractionIDs []uint64 `json:"attraction_ids"`
}
```

### packageInput — package.go

```go
type packageInput struct {
	Name                    string                `json:"name"`
	Description             string                `json:"description"`
	DurationDays            int                   `json:"duration_days"`
	DurationNights          int                   `json:"duration_nights"`
	Difficulty              string                `json:"difficulty"`
	NationalPriceCents      int                   `json:"national_price_cents"`
	ForeignSurchargeCents   int                   `json:"foreign_surcharge_cents"`
	Includes                []string              `json:"includes"`
	Excludes                []string              `json:"excludes"`
	Bring                   []string              `json:"bring"`
	CancellationAllowed     bool                  `json:"cancellation_allowed"`
	CancellationNoticeHours int                   `json:"cancellation_notice_hours"`
	Published               bool                  `json:"published"`
	Version                 int                   `json:"version"`
	Photos                  []packagePhotoInput   `json:"photos"`
	Itinerary               []packageDayInput     `json:"itinerary"`
	Schedule                *packageScheduleInput `json:"schedule"`
}
```

### packageResponse — package.go

```go
type packageResponse struct {
	models.TourPackage
	AgencyName       string `json:"agency_name"`
	MinimumPayingAge int    `json:"minimum_paying_age"`
}
```

### publicPackageDeparture — package_catalog.go

```go
type publicPackageDeparture struct {
	ID                  uint64    `json:"id"`
	StartsAt            time.Time `json:"starts_at"`
	MeetingAt           time.Time `json:"meeting_at"`
	BookingOpensAt      time.Time `json:"booking_opens_at"`
	BookingClosesAt     time.Time `json:"booking_closes_at"`
	MinCapacity         int       `json:"min_capacity"`
	MaxCapacity         int       `json:"max_capacity"`
	ConfirmedCapacity   int       `json:"confirmed_capacity"`
	AvailableCapacity   int       `json:"available_capacity"`
	RemainingForMinimum int       `json:"remaining_for_minimum"`
	MinimumReached      bool      `json:"minimum_reached"`
	MeetingPoint        string    `json:"meeting_point"`
	Instructions        string    `json:"instructions"`
	Status              string    `json:"status"`
	Bookable            bool      `json:"bookable"`
}
```

### publicPackageSummary — package_catalog.go

```go
type publicPackageSummary struct {
	ID                    uint64                    `json:"id"`
	AgencyID              uint64                    `json:"agency_id"`
	AgencyName            string                    `json:"agency_name"`
	AgencyDepartment      string                    `json:"agency_department"`
	AgencyCity            string                    `json:"agency_city"`
	Name                  string                    `json:"name"`
	Description           string                    `json:"description"`
	DurationDays          int                       `json:"duration_days"`
	DurationNights        int                       `json:"duration_nights"`
	Difficulty            string                    `json:"difficulty"`
	NationalPriceCents    int                       `json:"national_price_cents"`
	ForeignSurchargeCents int                       `json:"foreign_surcharge_cents"`
	Photos                []models.TourPackagePhoto `json:"photos"`
	FrequencyType         string                    `json:"frequency_type"`
	NextDeparture         *publicPackageDeparture   `json:"next_departure,omitempty"`
	Bookable              bool                      `json:"bookable"`
}
```

### publicPackageDetail — package_catalog.go

```go
type publicPackageDetail struct {
	publicPackageSummary
	Includes                []string                     `json:"includes"`
	Excludes                []string                     `json:"excludes"`
	Bring                   []string                     `json:"bring"`
	CancellationAllowed     bool                         `json:"cancellation_allowed"`
	CancellationNoticeHours int                          `json:"cancellation_notice_hours"`
	MinimumPayingAge        int                          `json:"minimum_paying_age"`
	Itinerary               []models.PackageItineraryDay `json:"itinerary"`
	Departures              []publicPackageDeparture     `json:"departures"`
}
```

### departureUpdateInput — package_departure.go

```go
type departureUpdateInput struct {
	Version            int    `json:"version"`
	MeetingTime        string `json:"meeting_time"`
	MinCapacity        int    `json:"min_capacity"`
	MaxCapacity        int    `json:"max_capacity"`
	BookingCutoffHours int    `json:"booking_cutoff_hours"`
	MeetingPoint       string `json:"meeting_point"`
	Instructions       string `json:"instructions"`
}
```

### departureCancelInput — package_departure.go

```go
type departureCancelInput struct {
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}
```

### minimumRefundInput — package_departure.go

```go
type minimumRefundInput struct {
	Version int `json:"version"`
}
```

### packageScheduleInput — package_schedule.go

```go
type packageScheduleInput struct {
	ID                  uint64 `json:"id"`
	FrequencyType       string `json:"frequency_type"`
	ValidFrom           string `json:"valid_from"`
	ValidUntil          string `json:"valid_until"`
	Weekdays            []int  `json:"weekdays"`
	DepartureTime       string `json:"departure_time"`
	MeetingTime         string `json:"meeting_time"`
	DefaultMinCapacity  int    `json:"default_min_capacity"`
	DefaultMaxCapacity  int    `json:"default_max_capacity"`
	BookingCutoffHours  int    `json:"booking_cutoff_hours"`
	MaximumAdvanceDays  int    `json:"maximum_advance_days"`
	DefaultMeetingPoint string `json:"default_meeting_point"`
	DefaultInstructions string `json:"default_instructions"`
}
```

### updateProfileRequest — profile.go

```go
type updateProfileRequest struct {
	FirstName      string `json:"first_name" validate:"required,min=2,max=100"`
	LastName       string `json:"last_name" validate:"required,min=2,max=100"`
	Phone          string `json:"phone" validate:"omitempty,max=30"`
	DocumentNumber string `json:"document_number" validate:"omitempty,max=40"`
	Nationality    string `json:"nationality" validate:"omitempty,max=80"`
}
```

### purchaseMinorInput — purchase.go

```go
type purchaseMinorInput struct {
	Age       int  `json:"age"`
	IsForeign bool `json:"is_foreign"`
}
```

### purchaseCreateInput — purchase.go

```go
type purchaseCreateInput struct {
	DepartureID        uint64               `json:"departure_id"`
	PaymentMethod      string               `json:"payment_method"`
	NationalAdults     int                  `json:"national_adults"`
	ForeignAdults      int                  `json:"foreign_adults"`
	Minors             []purchaseMinorInput `json:"minors"`
	PaymentProof       string               `json:"payment_proof"`
	ExpectedTotalCents *int                 `json:"expected_total_cents"`
}
```

### purchaseReviewInput — purchase.go

```go
type purchaseReviewInput struct {
	Version  int    `json:"version"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}
```

### purchaseProofInput — purchase.go

```go
type purchaseProofInput struct {
	Version      int    `json:"version"`
	PaymentProof string `json:"payment_proof"`
}
```

### refundDestinationInput — purchase.go

```go
type refundDestinationInput struct {
	Version             int    `json:"version"`
	RefundMethod        string `json:"refund_method"`
	RefundQR            string `json:"refund_qr"`
	RefundBankName      string `json:"refund_bank_name"`
	RefundAccountHolder string `json:"refund_account_holder"`
	RefundAccountNumber string `json:"refund_account_number"`
}
```

### purchaseCancelInput — purchase.go

```go
type purchaseCancelInput struct {
	refundDestinationInput
	Reason string `json:"reason"`
}
```

### refundCompleteInput — purchase.go

```go
type refundCompleteInput struct {
	Version         int    `json:"version"`
	RefundProof     string `json:"refund_proof"`
	RefundReference string `json:"refund_reference"`
}
```

### purchasePaymentOptions — purchase.go

```go
type purchasePaymentOptions struct {
	PackageID           uint64   `json:"package_id"`
	AgencyName          string   `json:"agency_name"`
	AgencyPhone         string   `json:"agency_phone,omitempty"`
	AgencyEmail         string   `json:"agency_email,omitempty"`
	MinimumPayingAge    int      `json:"minimum_paying_age"`
	Methods             []string `json:"methods"`
	BankName            string   `json:"bank_name"`
	AccountHolder       string   `json:"account_holder"`
	AccountNumber       string   `json:"account_number"`
	PaymentInstructions string   `json:"payment_instructions"`
	QRImage             string   `json:"qr_image,omitempty"`
}
```

### purchaseResponse — purchase.go

```go
type purchaseResponse struct {
	models.TourPackagePurchase
	PackageName          string     `json:"package_name"`
	AgencyName           string     `json:"agency_name"`
	AgencyPhone          string     `json:"agency_phone,omitempty"`
	AgencyEmail          string     `json:"agency_email,omitempty"`
	DepartureStart       time.Time  `json:"departure_start"`
	MeetingPoint         string     `json:"meeting_point"`
	HasProof             bool       `json:"has_proof"`
	HasRefundQR          bool       `json:"has_refund_qr"`
	HasRefundProof       bool       `json:"has_refund_proof"`
	Cancellable          bool       `json:"cancellable"`
	CancellationDeadline *time.Time `json:"cancellation_deadline,omitempty"`
	RefundOverdue        bool       `json:"refund_overdue"`
}
```

## Compra idempotente

POST /api/v1/me/purchases: `departure_id`, `payment_method`, `national_adults`, `foreign_adults`, `minors` (age/is_foreign), `payment_proof` y `expected_total_cents` opcional. Cabecera `Idempotency-Key` opcional de 16–128 caracteres alfanuméricos, guion o guion bajo. 201 para nueva compra, 200 para reintento idéntico, 409 PURCHASE_REQUEST_CONFLICT si cambia el contenido y 409 PURCHASE_PRICE_CHANGED si cambia el precio. Compra, retención de cupos y clave se confirman en una transacción. El pago es externo y su revisión humana.

## Errores comunes

400 validación; 401 sesión ausente/vencida; 403 rol/CSRF/origen; 404 recurso inexistente/no accesible; 409 conflicto de estado/versión/precio; 413 cuerpo demasiado grande; 429 límite de solicitudes; 500 error interno. Los estados exactos por handler deben consultarse en la fuente enlazada; no se afirma que cada ruta emita todos ellos.
