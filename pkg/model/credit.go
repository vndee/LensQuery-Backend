package model

import (
	"time"

	"gorm.io/gorm"
)

type UserCredits struct {
	*gorm.Model

	UserID               string `json:"user_id" gorm:"primaryKey"`
	PurchasedTimestampMs int64  `json:" purchased_timestamp_ms"`
	ExpiredTimestampMs   int64  `json:"expired_timestamp_ms"`
	AmmountEquationSnap  int    `json:"ammount_equation_snap"`
	RemainEquationSnap   int    `json:"remain_equation_snap"`
	AmmountTextSnap      int    `json:"ammount_text_snap"`
	RemainTextSnap       int    `json:"remain_text_snap"`
}

type CreditUsageHistory struct {
	*gorm.Model

	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Timestamp   time.Time `json:"timestamp"`
	RequestType string    `json:"request_type"`
}

type UserTrialData struct {
	*gorm.Model

	UserID             string `json:"user_id" gorm:"primaryKey"`
	Email              string `json:"email"`
	ExpiredTimestampMs int64  `json:"expired_timestamp_ms"`
}
