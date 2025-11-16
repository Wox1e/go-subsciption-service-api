package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServiceName string     `gorm:"type:varchar(255);not null;index" json:"service_name"`
	Price       int        `gorm:"type:integer;not null;check:price > 0" json:"price"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	StartDate   time.Time `gorm:"type:date;not null;index" json:"start_date"`
	EndDate     *time.Time `gorm:"type:date;index" json:"end_date,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int    `json:"price" binding:"required,min=1" example:"400"`
	UserID      string `json:"user_id" binding:"required,uuid" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" example:"12-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" example:"Yandex Plus"`
	Price       *int    `json:"price,omitempty" example:"400"`
	StartDate   *string `json:"start_date,omitempty" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type SubscriptionResponse struct {
	ID          string  `json:"id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"2025-07-01T00:00:00Z"`
	EndDate     *string `json:"end_date,omitempty" example:"2025-12-01T00:00:00Z"`
	CreatedAt   string  `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt   string  `json:"updated_at" example:"2025-01-01T00:00:00Z"`
}

type TotalCostRequest struct {
	UserID      *string `json:"user_id,omitempty" form:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName *string `json:"service_name,omitempty" form:"service_name" example:"Yandex Plus"`
	StartPeriod string  `json:"start_period" form:"start_period" binding:"required" example:"01-2025"`
	EndPeriod   string  `json:"end_period" form:"end_period" binding:"required" example:"12-2025"`
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost" example:"4800"`
}

