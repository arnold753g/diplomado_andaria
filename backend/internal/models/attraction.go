package models

import "time"

type AttractionCategory struct {
	ID            uint64                  `gorm:"primaryKey" json:"id"`
	Name          string                  `json:"name"`
	Position      int                     `json:"-"`
	Subcategories []AttractionSubcategory `gorm:"foreignKey:CategoryID" json:"subcategories"`
}
type AttractionSubcategory struct {
	ID         uint64 `gorm:"primaryKey" json:"id"`
	CategoryID uint64 `json:"category_id"`
	Name       string `json:"name"`
	Position   int    `json:"-"`
}
type AttractionClassification struct {
	AttractionID  uint64                `gorm:"primaryKey" json:"-"`
	SubcategoryID uint64                `gorm:"primaryKey" json:"subcategory_id"`
	Position      int                   `json:"position"`
	Subcategory   AttractionSubcategory `gorm:"foreignKey:SubcategoryID" json:"subcategory"`
}

type Attraction struct {
	Subcategories    []AttractionClassification `gorm:"foreignKey:AttractionID" json:"subcategories"`
	ScheduleMode     string                     `json:"schedule_mode"`
	OpeningTime      string                     `json:"opening_time"`
	ClosingTime      string                     `json:"closing_time"`
	OpeningDays      []int                      `gorm:"serializer:json" json:"opening_days"`
	SeasonMode       string                     `json:"season_mode"`
	SeasonStartMonth *int                       `json:"season_start_month"`
	SeasonEndMonth   *int                       `json:"season_end_month"`
	ID               uint64                     `gorm:"primaryKey" json:"id"`
	Name             string                     `json:"name"`
	Description      string                     `json:"description"`
	Category         string                     `json:"category"`
	Department       string                     `json:"department"`
	City             string                     `json:"city"`
	Address          string                     `json:"address"`
	Latitude         *float64                   `json:"latitude"`
	Longitude        *float64                   `json:"longitude"`
	OpeningHours     string                     `json:"opening_hours"`
	AdmissionCents   int                        `json:"admission_cents"`
	Recommendations  string                     `json:"recommendations"`
	Phone            string                     `json:"phone"`
	ManagerID        uint64                     `json:"-"`
	Manager          *User                      `gorm:"foreignKey:ManagerID" json:"-"`
	Status           string                     `json:"-"`
	Published        bool                       `json:"-"`
	Version          int                        `json:"-"`
	Photos           []AttractionPhoto          `gorm:"foreignKey:AttractionID" json:"photos"`
	CreatedAt        time.Time                  `json:"-"`
	UpdatedAt        time.Time                  `json:"updated_at"`
}

type AttractionPhoto struct {
	ID           uint64 `gorm:"primaryKey" json:"id"`
	AttractionID uint64 `json:"-"`
	Position     int    `json:"position"`
	Image        []byte `json:"-"`
}

type AttractionFavorite struct {
	UserID       uint64    `gorm:"primaryKey" json:"-"`
	AttractionID uint64    `gorm:"primaryKey" json:"attraction_id"`
	CreatedAt    time.Time `json:"created_at"`
}
