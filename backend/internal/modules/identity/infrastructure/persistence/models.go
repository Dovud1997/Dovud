package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	Email           string         `gorm:"size:255;not null"`
	Phone           *string        `gorm:"size:64"`
	PasswordHash    string         `gorm:"column:password_hash;not null"`
	FullName        string         `gorm:"size:255;not null"`
	Status          string         `gorm:"size:32;not null"`
	Locale          string         `gorm:"size:8;not null"`
	ThemePreference string         `gorm:"size:16;not null"`
	LastLoginAt     *time.Time
	IsPlatformAdmin bool           `gorm:"not null;default:false"`
	Version         int64          `gorm:"not null;default:1"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string { return "users" }

type RoleModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  *uuid.UUID     `gorm:"type:uuid;index"`
	Code      string         `gorm:"size:64;not null"`
	Name      string         `gorm:"size:128;not null"`
	IsSystem  bool           `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (RoleModel) TableName() string { return "roles" }

type PermissionModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code        string    `gorm:"size:128;uniqueIndex;not null"`
	Description string    `gorm:"size:255"`
	CreatedAt   time.Time
}

func (PermissionModel) TableName() string { return "permissions" }

type RolePermissionModel struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (RolePermissionModel) TableName() string { return "role_permissions" }

type UserRoleModel struct {
	UserID   uuid.UUID  `gorm:"type:uuid;primaryKey"`
	RoleID   uuid.UUID  `gorm:"type:uuid;primaryKey"`
	BranchID *uuid.UUID `gorm:"type:uuid"`
}

func (UserRoleModel) TableName() string { return "user_roles" }

type RefreshTokenModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null"`
	TokenHash string    `gorm:"size:128;uniqueIndex;not null"`
	DeviceID  *string   `gorm:"size:128"`
	ExpiresAt time.Time `gorm:"not null"`
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }

type UserDeviceModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	DeviceID   string    `gorm:"size:128;not null"`
	Platform   string    `gorm:"size:32;not null"`
	PushToken  *string
	AppVersion *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (UserDeviceModel) TableName() string { return "user_devices" }
