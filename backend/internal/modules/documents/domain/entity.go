package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	FileStatusPending   = "pending"
	FileStatusReady     = "ready"
	FileStatusFailed    = "failed"
	FileStatusDeleted   = "deleted"

	DocStatusDraft    = "draft"
	DocStatusActive   = "active"
	DocStatusArchived = "archived"
)

type File struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Bucket        string
	ObjectKey     string
	FileName      string
	Mime          string
	Size          int64
	Checksum      *string
	Status        string
	UploadedBy    *uuid.UUID
	ThumbnailKey  *string
	MetaJSON      *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
}

type Document struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Title       string
	Description *string
	DocType     string
	Status      string
	CreatedBy   *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DocumentFile struct {
	DocumentID uuid.UUID
	FileID     uuid.UUID
	Role       string
	CreatedAt  time.Time
}

type EntityFile struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	EntityType string
	EntityID   uuid.UUID
	FileID     uuid.UUID
	Role       string
	CreatedAt  time.Time
}

func ValidFileStatus(s string) bool {
	switch s {
	case FileStatusPending, FileStatusReady, FileStatusFailed, FileStatusDeleted:
		return true
	default:
		return false
	}
}
