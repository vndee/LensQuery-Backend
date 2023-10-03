package model

import (
	"time"

	"gorm.io/gorm"
)

type UserCredits struct {
	*gorm.Model

	UserID               string  `json:"user_id" gorm:"primaryKey"`
	PurchasedTimestampMs int64   `json:"purchased_timestamp_ms"`
	CreditAmount         float64 `json:"credit_amount"`
}

type CreditUsageHistory struct {
	*gorm.Model

	UserID       string    `json:"user_id"`
	Amount       float64   `json:"amount"`
	Timestamp    time.Time `json:"timestamp"`
	RequestType  string    `json:"request_type"`
	GenerationID string    `json:"generation_id"`
}

type UserTrialData struct {
	*gorm.Model

	UserID             string `json:"user_id" gorm:"primaryKey"`
	Email              string `json:"email"`
	ExpiredTimestampMs int64  `json:"expired_timestamp_ms"`
}

type Receipt struct {
	*gorm.Model

	ID                     string  `json:"id" gorm:"primaryKey"`
	ModelType              string  `json:"model"`
	Streamed               bool    `json:"streamed"`
	GenerationTime         float64 `json:"generation_time"`
	CreatedAt              string  `json:"created_at"`
	TokensPrompt           float64 `json:"tokens_prompt"`
	TokensCompletion       float64 `json:"tokens_completion"`
	NativeTokensPrompt     float64 `json:"native_tokens_prompt"`
	NativeTokensCompletion float64 `json:"native_tokens_completion"`
	NumMediaGenerations    float64 `json:"num_media_generations"`
	Origin                 string  `json:"origin"`
	Usage                  float64 `json:"usage"`
}

type ReceiptResponse struct {
	Data Receipt `json:"data"`
}
