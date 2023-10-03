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
	OpenRouterAPIKey   = "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"

	// Pricing
	PriceAdjustFactor     = 1.1
	MinPrice              = 0.001
	FreeTextSnapPrice     = 0.01
	EquationTextSnapPrice = 0.02
)
