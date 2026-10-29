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
	ID                  uint64     `gorm:"column:id;primaryKey" json:"id"`
	Email               string     `gorm:"column:correo;size:320;not null;uniqueIndex" json:"email"`
	PasswordHash        string     `gorm:"column:hash_contrasena;size:255;not null" json:"-"`
	FirstName           string     `gorm:"column:nombres;size:100;not null" json:"first_name"`
	LastName            string     `gorm:"column:apellidos;size:100;not null" json:"last_name"`
	Phone               string     `gorm:"column:telefono;size:30;not null" json:"phone"`
	DocumentNumber      string     `gorm:"column:numero_documento;size:40;not null" json:"document_number"`
	Nationality         string     `gorm:"column:nacionalidad;size:80;not null" json:"nationality"`
	GoogleSubject       *string    `gorm:"column:identificador_google;size:255;uniqueIndex" json:"-"`
	EmailVerifiedAt     *time.Time `gorm:"column:correo_verificado_en" json:"-"`
	Role                string     `gorm:"column:rol;size:20;not null" json:"role"`
	Status              string     `gorm:"column:estado;size:20;not null" json:"status"`
	FailedLoginAttempts int        `gorm:"column:intentos_fallidos_acceso;not null;default:0" json:"-"`
	LockedUntil         *time.Time `gorm:"column:bloqueado_hasta" json:"-"`
	LastLoginAt         *time.Time `gorm:"column:ultimo_acceso_en" json:"last_login_at"`
	CreatedAt           time.Time  `gorm:"column:creado_en" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:actualizado_en" json:"updated_at"`
}

func (User) TableName() string { return "usuarios" }

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
