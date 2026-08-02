package persistence

import (
	"time"

	"github.com/google/uuid"
)

type NotificationModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Type        string     `gorm:"size:64;not null"`
	Title       string     `gorm:"size:255;not null"`
	Body        string     `gorm:"type:text;not null"`
	PayloadJSON *string    `gorm:"type:text"`
	ReadAt      *time.Time `gorm:"index"`
	Channel     string     `gorm:"size:32;not null;index"`
	CreatedAt   time.Time  `gorm:"index"`
}

func (NotificationModel) TableName() string { return "notifications" }

type NotificationDeliveryModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	NotificationID uuid.UUID `gorm:"type:uuid;not null;index"`
	Channel        string    `gorm:"size:32;not null"`
	Status         string    `gorm:"size:32;not null;index"`
	Error          *string   `gorm:"type:text"`
	AttemptedAt    time.Time `gorm:"not null"`
}

func (NotificationDeliveryModel) TableName() string { return "notification_deliveries" }
