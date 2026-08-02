package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	Bucket      string         `gorm:"size:128;not null"`
	ObjectKey   string         `gorm:"size:512;not null"`
	FileName    string         `gorm:"size:255;not null"`
	Mime        string         `gorm:"size:128;not null"`
	Size        int64          `gorm:"not null;default:0"`
	Checksum    *string        `gorm:"size:128"`
	Status       string         `gorm:"size:32;not null;index"`
	UploadedBy   *uuid.UUID     `gorm:"type:uuid"`
	ThumbnailKey *string        `gorm:"size:512"`
	MetaJSON     *string        `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (FileModel) TableName() string { return "files" }

type DocumentModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	Title       string         `gorm:"size:255;not null"`
	Description *string        `gorm:"type:text"`
	DocType     string         `gorm:"size:64;not null;default:general"`
	Status      string         `gorm:"size:32;not null;default:active"`
	CreatedBy   *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (DocumentModel) TableName() string { return "documents" }

type DocumentFileModel struct {
	DocumentID uuid.UUID `gorm:"type:uuid;primaryKey"`
	FileID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role       string    `gorm:"size:64;not null;default:attachment"`
	CreatedAt  time.Time
}

func (DocumentFileModel) TableName() string { return "document_files" }

type EntityFileModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index"`
	EntityType string    `gorm:"size:64;not null;index"`
	EntityID   uuid.UUID `gorm:"type:uuid;not null;index"`
	FileID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Role       string    `gorm:"size:64;not null;default:attachment"`
	CreatedAt  time.Time
}

func (EntityFileModel) TableName() string { return "entity_files" }
