package model

import "time"

type UserCredits struct {
	UserID  string  `json:"user_id"`
	Credits float64 `json:"credits"`
}

type CreditUsageHistory struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Timestamp   time.Time `json:"timestamp"`
	RequestType string    `json:"request_type"`
}
