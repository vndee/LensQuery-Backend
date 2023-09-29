package model

import "gorm.io/gorm"

type VerificationCode struct {
	*gorm.Model

	Type  string `json:"type"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

type RequestResetPasswordParams struct {
	Recipient string `json:"recipient"`
}

type VerifyResetPasswordParams struct {
	Type  string `json:"type"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResetPasswordParams struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type DeleteAccountParams struct {
	UserId string `json:"user_id"`
}

type ActivateUserTrialParams struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
}

type CheckTrialPlanParams struct {
	UserId string `json:"user_id"`
}
