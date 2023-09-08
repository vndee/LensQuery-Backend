package model

import (
	"time"

	"gorm.io/gorm"
)

type UserCredits struct {
	gorm.Model

	UserID  string  `json:"user_id"`
	Credits float64 `json:"credits"`
}

type CreditUsageHistory struct {
	gorm.Model

	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Timestamp   time.Time `json:"timestamp"`
	RequestType string    `json:"request_type"`
}
