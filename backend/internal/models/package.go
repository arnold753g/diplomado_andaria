package models

import "time"

type TourPackage struct {
	ID                      uint64                 `gorm:"primaryKey" json:"id"`
	AgencyID                uint64                 `json:"agency_id"`
	Agency                  *Agency                `gorm:"foreignKey:AgencyID" json:"-"`
	Name                    string                 `json:"name"`
	Description             string                 `json:"description"`
	DurationDays            int                    `json:"duration_days"`
	DurationNights          int                    `json:"duration_nights"`
	Difficulty              string                 `json:"difficulty"`
	NationalPriceCents      int                    `json:"national_price_cents"`
	ForeignSurchargeCents   int                    `json:"foreign_surcharge_cents"`
	Includes                []string               `gorm:"serializer:json" json:"includes"`
	Excludes                []string               `gorm:"serializer:json" json:"excludes"`
	Bring                   []string               `gorm:"serializer:json" json:"bring"`
	CancellationAllowed     bool                   `json:"cancellation_allowed"`
	CancellationNoticeHours int                    `json:"cancellation_notice_hours"`
	Published               bool                   `json:"published"`
	Version                 int                    `json:"version"`
	Photos                  []TourPackagePhoto     `gorm:"foreignKey:PackageID" json:"photos"`
	Itinerary               []PackageItineraryDay  `gorm:"foreignKey:PackageID" json:"itinerary"`
	Schedule                *TourPackageSchedule   `gorm:"foreignKey:PackageID" json:"schedule,omitempty"`
	Departures              []TourPackageDeparture `gorm:"foreignKey:PackageID" json:"departures,omitempty"`
	CreatedAt               time.Time              `json:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at"`
}

func (TourPackage) TableName() string { return "tour_packages" }

type TourPackagePhoto struct {
	ID        uint64 `gorm:"primaryKey" json:"id"`
	PackageID uint64 `json:"-"`
	Position  int    `json:"position"`
	Image     []byte `json:"-"`
}

func (TourPackagePhoto) TableName() string { return "tour_package_photos" }

type PackageItineraryDay struct {
	ID          uint64                       `gorm:"primaryKey" json:"id"`
	PackageID   uint64                       `json:"-"`
	DayNumber   int                          `json:"day_number"`
	Title       string                       `json:"title"`
	Description string                       `json:"description"`
	Activities  []string                     `gorm:"serializer:json" json:"activities"`
	Attractions []PackageItineraryAttraction `gorm:"foreignKey:ItineraryDayID" json:"attractions"`
}

func (PackageItineraryDay) TableName() string { return "tour_package_itinerary_days" }

type PackageItineraryAttraction struct {
	ItineraryDayID uint64      `gorm:"primaryKey" json:"-"`
	AttractionID   uint64      `gorm:"primaryKey" json:"attraction_id"`
	Position       int         `json:"position"`
	Attraction     *Attraction `gorm:"foreignKey:AttractionID" json:"attraction,omitempty"`
}

func (PackageItineraryAttraction) TableName() string { return "tour_package_itinerary_attractions" }

type TourPackageSchedule struct {
	ID                  uint64    `gorm:"primaryKey" json:"id"`
	PackageID           uint64    `json:"-"`
	FrequencyType       string    `json:"frequency_type"`
	ValidFrom           time.Time `gorm:"type:date" json:"valid_from"`
	ValidUntil          time.Time `gorm:"type:date" json:"valid_until"`
	Weekdays            []int     `gorm:"serializer:json" json:"weekdays"`
	DepartureTime       string    `gorm:"type:time" json:"departure_time"`
	MeetingTime         string    `gorm:"type:time" json:"meeting_time"`
	DefaultMinCapacity  int       `json:"default_min_capacity"`
	DefaultMaxCapacity  int       `json:"default_max_capacity"`
	BookingCutoffHours  int       `json:"booking_cutoff_hours"`
	MaximumAdvanceDays  int       `json:"maximum_advance_days"`
	DefaultMeetingPoint string    `json:"default_meeting_point"`
	DefaultInstructions string    `json:"default_instructions"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (TourPackageSchedule) TableName() string { return "tour_package_schedules" }

type TourPackageDeparture struct {
	ID                  uint64     `gorm:"primaryKey" json:"id"`
	PackageID           uint64     `json:"package_id"`
	ScheduleID          *uint64    `json:"schedule_id,omitempty"`
	StartsAt            time.Time  `json:"starts_at"`
	MeetingAt           time.Time  `json:"meeting_at"`
	BookingOpensAt      time.Time  `json:"booking_opens_at"`
	BookingClosesAt     time.Time  `json:"booking_closes_at"`
	MinCapacity         int        `json:"min_capacity"`
	MaxCapacity         int        `json:"max_capacity"`
	HeldCapacity        int        `json:"held_capacity"`
	ConfirmedCapacity   int        `json:"confirmed_capacity"`
	MeetingPoint        string     `json:"meeting_point"`
	Instructions        string     `json:"instructions"`
	Status              string     `json:"status"`
	CancellationReason  string     `json:"cancellation_reason"`
	IsException         bool       `json:"is_exception"`
	ModifiedByUserID    *uint64    `json:"modified_by_user_id,omitempty"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	MinimumReviewAt     *time.Time `json:"minimum_review_at,omitempty"`
	MinimumReviewedAt   *time.Time `json:"minimum_reviewed_at,omitempty"`
	MinimumReviewUserID *uint64    `json:"minimum_review_user_id,omitempty"`
	Version             int        `json:"version"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (TourPackageDeparture) TableName() string { return "tour_package_departures" }
