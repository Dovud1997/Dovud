package syncport

import (
	"context"

	"github.com/google/uuid"
)

// ApplyRequest is a sync push / conflict-resolve mutation against domain tables.
type ApplyRequest struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	EntityType string
	EntityID   string
	Op         string // create | update | delete
	Payload    map[string]any
}

// ApplyResult is the canonical domain snapshot after apply.
type ApplyResult struct {
	Payload any
	Version int64
	Deleted bool
}

// EntityApplicator mutates domain tables for syncable entity types.
type EntityApplicator interface {
	Apply(ctx context.Context, req ApplyRequest) (*ApplyResult, error)
	Supports(entityType string) bool
}
