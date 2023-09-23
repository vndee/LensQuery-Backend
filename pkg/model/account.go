package model

import "gorm.io/gorm"

type VerificationCode struct {
	*gorm.Model

	Type  string `json:"type"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResetPasswordParams struct {
	UserId      string `json:"user_id"`
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}
