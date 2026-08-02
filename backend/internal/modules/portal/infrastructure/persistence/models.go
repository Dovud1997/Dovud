package persistence

import (
	"time"

	"github.com/google/uuid"
)

type CustomerUserModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_customer_users_user"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time
}

func (CustomerUserModel) TableName() string { return "customer_users" }
