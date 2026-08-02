package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	ChannelInApp = "in_app"
	ChannelPush  = "push"
	ChannelEmail = "email"
	ChannelSMS   = "sms"

	DeliveryPending = "pending"
	DeliverySent    = "sent"
	DeliveryFailed  = "failed"
)

type Notification struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	UserID      uuid.UUID
	Type        string
	Title       string
	Body        string
	PayloadJSON *string
	ReadAt      *time.Time
	Channel     string
	CreatedAt   time.Time
}

type NotificationDelivery struct {
	ID             uuid.UUID
	NotificationID uuid.UUID
	Channel        string
	Status         string
	Error          *string
	DeviceID       *string
	Platform       *string
	TokenSuffix    *string
	AttemptedAt    time.Time
}

func ValidChannel(c string) bool {
	switch c {
	case ChannelInApp, ChannelPush, ChannelEmail, ChannelSMS:
		return true
	default:
		return false
	}
}

func ValidDeliveryStatus(s string) bool {
	switch s {
	case DeliveryPending, DeliverySent, DeliveryFailed:
		return true
	default:
		return false
	}
}
