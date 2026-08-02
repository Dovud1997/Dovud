package domain

import (
	"context"

	"github.com/google/uuid"
)

type ListFilters struct {
	UnreadOnly bool
}

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification, delivery *NotificationDelivery) error
	ListByUser(ctx context.Context, tenantID, userID uuid.UUID, filters ListFilters, page, perPage int) ([]Notification, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Notification, error)
	MarkRead(ctx context.Context, tenantID, userID, id uuid.UUID) error
	MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int64, error)
	UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int64, error)
	ListDeliveries(ctx context.Context, notificationID uuid.UUID) ([]NotificationDelivery, error)
	UpdateDeliveryStatus(ctx context.Context, notificationID uuid.UUID, channel, status string, errMsg *string) error
}
