package config

import "time"

const (
	AccountVerificationCodeTTL = 10 * time.Minute
	EmailLimiterRate           = 5
	EmailLimiterBurst          = 1
	EmailLimiterPeriod         = 10 * time.Minute
	IPLimiterRate              = 5
	IPLimiterBurst             = 1
	IPLimiterPeriod            = 10 * time.Minute
)
