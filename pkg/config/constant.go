package config

import (
	"time"
)

const (
	// Rate limiter
	AccountVerificationCodeTTL = 10 * time.Minute
	EmailLimiterRate           = 5
	EmailLimiterBurst          = 1
	EmailLimiterPeriod         = 10 * time.Minute
	IPLimiterRate              = 5
	IPLimiterBurst             = 1
	IPLimiterPeriod            = 10 * time.Minute

	// Trial period
	TrialPeriod              = 7 * 24 * time.Hour
	TrialFreeTextSnapCredits = 30
	TrialFreeEquationCredits = 20
	TrialCreditAmount        = 0.1

	// Open Router
	OpenRouterEndpoint = "https://openrouter.ai/api/v1"
	OpenRouterAPIKey   = "no-key"

	// Pricing
	PriceAdjustFactor     = 1.1
	MinPrice              = 0.001
	FreeTextSnapPrice     = 0.01
	EquationTextSnapPrice = 0.02
)
