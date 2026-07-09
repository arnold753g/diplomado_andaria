package models

import "time"

const (
	RoleAdmin      = "admin"
	RoleUser       = "turista"
	RoleAgency     = "encargado_agencia"
	RoleAttraction = "encargado_atraccion"

	StatusActive   = "active"
	StatusInactive = "inactive"
)

type User struct {
	ID                  uint64     `gorm:"primaryKey" json:"id"`
	Email               string     `gorm:"size:320;not null;uniqueIndex" json:"email"`
	PasswordHash        string     `gorm:"size:255;not null" json:"-"`
	FirstName           string     `gorm:"size:100;not null" json:"first_name"`
	LastName            string     `gorm:"size:100;not null" json:"last_name"`
	Phone               string     `gorm:"size:30;not null" json:"phone"`
	DocumentNumber      string     `gorm:"size:40;not null" json:"document_number"`
	Nationality         string     `gorm:"size:80;not null" json:"nationality"`
	GoogleSubject       *string    `gorm:"size:255;uniqueIndex" json:"-"`
	Role                string     `gorm:"size:20;not null" json:"role"`
	Status              string     `gorm:"size:20;not null" json:"status"`
	FailedLoginAttempts int        `gorm:"not null;default:0" json:"-"`
	LockedUntil         *time.Time `json:"-"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (User) TableName() string { return "users" }

type PublicUser struct {
	ID             uint64     `json:"id"`
	Email          string     `json:"email"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Phone          string     `json:"phone"`
	DocumentNumber string     `json:"document_number"`
	Nationality    string     `json:"nationality"`
	HasPassword    bool       `json:"has_password"`
	GoogleLinked   bool       `json:"google_linked"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (u User) Public() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email, FirstName: u.FirstName, LastName: u.LastName,
		Role: u.Role, Status: u.Status, LastLoginAt: u.LastLoginAt, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		Phone: u.Phone, DocumentNumber: u.DocumentNumber, Nationality: u.Nationality,
		HasPassword: u.PasswordHash != "", GoogleLinked: u.GoogleSubject != nil}
}

func ValidRole(role string) bool {
	return role == RoleAdmin || role == RoleUser || role == RoleAgency || role == RoleAttraction
}
