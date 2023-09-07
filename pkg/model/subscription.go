package model

import (
	"time"

	"gorm.io/gorm"
)

type SubcriptionPlan struct {
	gorm.Model

	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Credits     float64       `json:"credits"`
	Price       float64       `json:"price"`
	Duration    time.Duration `json:"duration"`
	Description string        `json:"description"`
}

type UserSubscription struct {
	gorm.Model

	ID                int64     `json:"id"`
	UserID            string    `json:"user_id"`
	SubcriptionPlanID string    `json:"subcription_plan_id"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	IsActive          bool      `json:"is_active"`
	Receipt           string    `json:"receipt"`
}
