package models

import "time"

type Agency struct {
	ID                  uint64    `gorm:"column:id;primaryKey" json:"id"`
	Name                string    `gorm:"column:nombre" json:"name"`
	Description         string    `gorm:"column:descripcion" json:"description"`
	Department          string    `gorm:"column:departamento" json:"department"`
	City                string    `gorm:"column:ciudad" json:"city"`
	Address             string    `gorm:"column:direccion" json:"address"`
	Phone               string    `gorm:"column:telefono" json:"phone"`
	Email               string    `gorm:"column:correo" json:"email"`
	ManagerID           uint64    `gorm:"column:encargado_id" json:"manager_id"`
	Manager             *User     `gorm:"foreignKey:ManagerID" json:"-"`
	Status              string    `gorm:"column:estado" json:"status"`
	Published           bool      `gorm:"column:publicado" json:"published"`
	MinimumPayingAge    int       `gorm:"column:edad_minima_pago" json:"minimum_paying_age"`
	AcceptsQR           bool      `gorm:"column:acepta_qr" json:"accepts_qr"`
	AcceptsTransfer     bool      `gorm:"column:acepta_transferencia" json:"accepts_transfer"`
	BankName            string    `gorm:"column:nombre_banco" json:"bank_name"`
	AccountHolder       string    `gorm:"column:titular_cuenta" json:"account_holder"`
	AccountNumber       string    `gorm:"column:numero_cuenta" json:"account_number"`
	PaymentInstructions string    `gorm:"column:instrucciones_pago" json:"payment_instructions"`
	QRImage             []byte    `gorm:"column:imagen_qr" json:"-"`
	Version             int       `gorm:"column:version" json:"version"`
	CreatedAt           time.Time `gorm:"column:creado_en" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:actualizado_en" json:"updated_at"`
}

func (Agency) TableName() string { return "agencias" }
