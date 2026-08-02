package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID          uuid.UUID
	TenantID    *uuid.UUID
	ActorUserID *uuid.UUID
	Action      string
	EntityType  *string
	EntityID    *uuid.UUID
	BeforeJSON  *string
	AfterJSON   *string
	IP          *string
	UserAgent   *string
	RequestID   *string
	CreatedAt   time.Time
}

type ListFilters struct {
	ActorUserID *uuid.UUID
	EntityType  string
	EntityID    *uuid.UUID
	Action      string
	From        *time.Time
	To          *time.Time
}
