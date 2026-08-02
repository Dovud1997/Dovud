package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	SyncProtocol = 1

	ConflictStatusOpen     = "open"
	ConflictStatusResolved = "resolved"

	ResolutionServerWins = "server_wins"
	ResolutionClientWins = "client_wins"

	OpCreate = "create"
	OpUpdate = "update"
	OpDelete = "delete"

	PushAcked    = "acked"
	PushConflict = "conflict"
	PushRejected = "rejected"
)

type SyncDevice struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	UserID         uuid.UUID
	DeviceID       string
	Platform       *string
	AppVersion     *string
	LastPullCursor string
	LastPushAt     *time.Time
	LastPullAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SyncChange struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	EntityType  string
	EntityID    string
	Version     int64
	Deleted     bool
	UpdatedAt   time.Time
	PayloadJSON string
}

type SyncConflict struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	DeviceID      string
	UserID        uuid.UUID
	EntityType    string
	EntityID      string
	ClientOpID    string
	BaseVersion   int64
	ServerVersion int64
	ClientPayload string
	ServerPayload string
	Status        string
	Resolution    *string
	CreatedAt     time.Time
	ResolvedAt    *time.Time
}

type SyncOp struct {
	OpID        string         `json:"op_id"`
	EntityType  string         `json:"entity_type"`
	EntityID    string         `json:"entity_id"`
	Op          string         `json:"op"`
	BaseVersion int64          `json:"base_version"`
	Payload     map[string]any `json:"payload"`
	ClientTS    time.Time      `json:"client_ts"`
}

func ValidOp(op string) bool {
	switch op {
	case OpCreate, OpUpdate, OpDelete:
		return true
	default:
		return false
	}
}
