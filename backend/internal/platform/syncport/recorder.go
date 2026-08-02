package syncport

import (
	"context"

	"github.com/google/uuid"
)

// ChangeRecorder fans domain writes into the sync change log for offline pull.
type ChangeRecorder interface {
	RecordChange(ctx context.Context, tenantID uuid.UUID, entityType, entityID string, version int64, deleted bool, payload any) error
}
