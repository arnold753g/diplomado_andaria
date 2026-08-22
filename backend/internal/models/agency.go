package models

import "time"

type Agency struct {
	ID                  uint64    `gorm:"primaryKey" json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Department          string    `json:"department"`
	City                string    `json:"city"`
	Address             string    `json:"address"`
	Phone               string    `json:"phone"`
	Email               string    `json:"email"`
	ManagerID           uint64    `json:"manager_id"`
	Manager             *User     `gorm:"foreignKey:ManagerID" json:"-"`
	Status              string    `json:"status"`
	Published           bool      `json:"published"`
	MinimumPayingAge    int       `json:"minimum_paying_age"`
	AcceptsQR           bool      `json:"accepts_qr"`
	AcceptsTransfer     bool      `json:"accepts_transfer"`
	BankName            string    `json:"bank_name"`
	AccountHolder       string    `json:"account_holder"`
	AccountNumber       string    `json:"account_number"`
	PaymentInstructions string    `json:"payment_instructions"`
	QRImage             []byte    `json:"-"`
	Version             int       `json:"version"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (Agency) TableName() string { return "agencies" }
