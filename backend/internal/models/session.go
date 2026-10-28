package models

import "time"

type Session struct {
	ID             string     `gorm:"column:id;primaryKey;size:64"`
	UserID         uint64     `gorm:"column:usuario_id;not null;index"`
	TokenHash      string     `gorm:"column:hash_token;size:64;not null;uniqueIndex"`
	LastActivityAt time.Time  `gorm:"column:ultima_actividad_en;not null;index"`
	ExpiresAt      time.Time  `gorm:"column:vence_en;not null;index"`
	UserAgent      string     `gorm:"column:agente_usuario;size:512"`
	IPAddress      string     `gorm:"column:direccion_ip;size:64"`
	RevokedAt      *time.Time `gorm:"column:revocado_en;index"`
	CreatedAt      time.Time  `gorm:"column:creado_en"`
	UpdatedAt      time.Time  `gorm:"column:actualizado_en"`
	User           User       `gorm:"foreignKey:UserID"`
}

func (Session) TableName() string { return "sesiones_autenticacion" }
