package models

import "time"

type TourPackage struct {
	ID                      uint64                 `gorm:"column:id;primaryKey" json:"id"`
	AgencyID                uint64                 `gorm:"column:agencia_id" json:"agency_id"`
	Agency                  *Agency                `gorm:"foreignKey:AgencyID" json:"-"`
	Name                    string                 `gorm:"column:nombre" json:"name"`
	Description             string                 `gorm:"column:descripcion" json:"description"`
	DurationDays            int                    `gorm:"column:duracion_dias" json:"duration_days"`
	DurationNights          int                    `gorm:"column:duracion_noches" json:"duration_nights"`
	Difficulty              string                 `gorm:"column:dificultad" json:"difficulty"`
	NationalPriceCents      int                    `gorm:"column:precio_nacional_centavos" json:"national_price_cents"`
	ForeignSurchargeCents   int                    `gorm:"column:recargo_extranjero_centavos" json:"foreign_surcharge_cents"`
	Includes                []string               `gorm:"column:incluye;serializer:json" json:"includes"`
	Excludes                []string               `gorm:"column:no_incluye;serializer:json" json:"excludes"`
	Bring                   []string               `gorm:"column:llevar;serializer:json" json:"bring"`
	CancellationAllowed     bool                   `gorm:"column:permite_cancelacion" json:"cancellation_allowed"`
	CancellationNoticeHours int                    `gorm:"column:horas_aviso_cancelacion" json:"cancellation_notice_hours"`
	Published               bool                   `gorm:"column:publicado" json:"published"`
	Version                 int                    `gorm:"column:version" json:"version"`
	Photos                  []TourPackagePhoto     `gorm:"foreignKey:PackageID" json:"photos"`
	Itinerary               []PackageItineraryDay  `gorm:"foreignKey:PackageID" json:"itinerary"`
	Schedule                *TourPackageSchedule   `gorm:"foreignKey:PackageID" json:"schedule,omitempty"`
	Departures              []TourPackageDeparture `gorm:"foreignKey:PackageID" json:"departures,omitempty"`
	CreatedAt               time.Time              `gorm:"column:creado_en" json:"created_at"`
	UpdatedAt               time.Time              `gorm:"column:actualizado_en" json:"updated_at"`
}

func (TourPackage) TableName() string { return "paquetes_turisticos" }

type TourPackagePhoto struct {
	ID        uint64 `gorm:"column:id;primaryKey" json:"id"`
	PackageID uint64 `gorm:"column:paquete_id" json:"-"`
	Position  int    `gorm:"column:posicion" json:"position"`
	Image     []byte `gorm:"column:imagen" json:"-"`
}

func (TourPackagePhoto) TableName() string { return "fotos_paquete" }

type PackageItineraryDay struct {
	ID          uint64                       `gorm:"column:id;primaryKey" json:"id"`
	PackageID   uint64                       `gorm:"column:paquete_id" json:"-"`
	DayNumber   int                          `gorm:"column:numero_dia" json:"day_number"`
	Title       string                       `gorm:"column:titulo" json:"title"`
	Description string                       `gorm:"column:descripcion" json:"description"`
	Activities  []string                     `gorm:"column:actividades;serializer:json" json:"activities"`
	Attractions []PackageItineraryAttraction `gorm:"foreignKey:ItineraryDayID" json:"attractions"`
}

func (PackageItineraryDay) TableName() string { return "dias_itinerario_paquete" }

type PackageItineraryAttraction struct {
	ItineraryDayID uint64      `gorm:"column:dia_itinerario_id;primaryKey" json:"-"`
	AttractionID   uint64      `gorm:"column:atraccion_id;primaryKey" json:"attraction_id"`
	Position       int         `gorm:"column:posicion" json:"position"`
	Attraction     *Attraction `gorm:"foreignKey:AttractionID" json:"attraction,omitempty"`
}

func (PackageItineraryAttraction) TableName() string { return "atracciones_itinerario_paquete" }

type TourPackageSchedule struct {
	ID                  uint64    `gorm:"column:id;primaryKey" json:"id"`
	PackageID           uint64    `gorm:"column:paquete_id" json:"-"`
	FrequencyType       string    `gorm:"column:tipo_frecuencia" json:"frequency_type"`
	ValidFrom           time.Time `gorm:"column:vigente_desde;type:date" json:"valid_from"`
	ValidUntil          time.Time `gorm:"column:vigente_hasta;type:date" json:"valid_until"`
	Weekdays            []int     `gorm:"column:dias_semana;serializer:json" json:"weekdays"`
	DepartureTime       string    `gorm:"column:hora_salida;type:time" json:"departure_time"`
	MeetingTime         string    `gorm:"column:hora_encuentro;type:time" json:"meeting_time"`
	DefaultMinCapacity  int       `gorm:"column:cupo_minimo_predeterminado" json:"default_min_capacity"`
	DefaultMaxCapacity  int       `gorm:"column:cupo_maximo_predeterminado" json:"default_max_capacity"`
	BookingCutoffHours  int       `gorm:"column:horas_cierre_compras" json:"booking_cutoff_hours"`
	MaximumAdvanceDays  int       `gorm:"column:dias_maxima_anticipacion" json:"maximum_advance_days"`
	DefaultMeetingPoint string    `gorm:"column:punto_encuentro_predeterminado" json:"default_meeting_point"`
	DefaultInstructions string    `gorm:"column:instrucciones_predeterminadas" json:"default_instructions"`
	CreatedAt           time.Time `gorm:"column:creado_en" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:actualizado_en" json:"updated_at"`
}

func (TourPackageSchedule) TableName() string { return "programaciones_paquete" }

type TourPackageDeparture struct {
	ID                            uint64     `gorm:"column:id;primaryKey" json:"id"`
	PackageID                     uint64     `gorm:"column:paquete_id" json:"package_id"`
	ScheduleID                    *uint64    `gorm:"column:programacion_id" json:"schedule_id,omitempty"`
	StartsAt                      time.Time  `gorm:"column:salida_en" json:"starts_at"`
	MeetingAt                     time.Time  `gorm:"column:encuentro_en" json:"meeting_at"`
	BookingOpensAt                time.Time  `gorm:"column:compras_abren_en" json:"booking_opens_at"`
	BookingClosesAt               time.Time  `gorm:"column:compras_cierran_en" json:"booking_closes_at"`
	MinCapacity                   int        `gorm:"column:cupo_minimo" json:"min_capacity"`
	MaxCapacity                   int        `gorm:"column:cupo_maximo" json:"max_capacity"`
	HeldCapacity                  int        `gorm:"column:cupos_reservados" json:"held_capacity"`
	ConfirmedCapacity             int        `gorm:"column:cupos_confirmados" json:"confirmed_capacity"`
	MeetingPoint                  string     `gorm:"column:punto_encuentro" json:"meeting_point"`
	Instructions                  string     `gorm:"column:instrucciones" json:"instructions"`
	Status                        string     `gorm:"column:estado" json:"status"`
	CancellationReason            string     `gorm:"column:motivo_cancelacion" json:"cancellation_reason"`
	IsException                   bool       `gorm:"column:es_excepcion" json:"is_exception"`
	ModifiedByUserID              *uint64    `gorm:"column:usuario_modificacion_id" json:"modified_by_user_id,omitempty"`
	CancelledAt                   *time.Time `gorm:"column:cancelado_en" json:"cancelled_at,omitempty"`
	MinimumReviewAt               *time.Time `gorm:"column:revision_minimo_en" json:"minimum_review_at,omitempty"`
	MinimumReviewedAt             *time.Time `gorm:"column:minimo_revisado_en" json:"minimum_reviewed_at,omitempty"`
	MinimumReviewUserID           *uint64    `gorm:"column:usuario_revision_minimo_id" json:"minimum_review_user_id,omitempty"`
	AuthorizedAt                  *time.Time `gorm:"column:autorizado_en" json:"authorized_at,omitempty"`
	AuthorizedByUserID            *uint64    `gorm:"column:usuario_autorizacion_id" json:"authorized_by_user_id,omitempty"`
	AuthorizationRevokedAt        *time.Time `gorm:"column:autorizacion_revocada_en" json:"authorization_revoked_at,omitempty"`
	AuthorizationRevocationReason string     `gorm:"column:motivo_revocacion_autorizacion" json:"authorization_revocation_reason"`
	Version                       int        `gorm:"column:version" json:"version"`
	CreatedAt                     time.Time  `gorm:"column:creado_en" json:"created_at"`
	UpdatedAt                     time.Time  `gorm:"column:actualizado_en" json:"updated_at"`
}

func (TourPackageDeparture) TableName() string { return "salidas_paquete" }
