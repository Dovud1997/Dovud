package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	Email           string
	Phone           *string
	PasswordHash    string
	FullName        string
	Status          string
	Locale          string
	ThemePreference string
	LastLoginAt     *time.Time
	IsPlatformAdmin bool
	Version         int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

func (u User) IsActive() bool { return u.Status == "active" && u.DeletedAt == nil }

type Role struct {
	ID        uuid.UUID
	TenantID  *uuid.UUID
	Code      string
	Name      string
	IsSystem  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Permission struct {
	ID          uuid.UUID
	Code        string
	Description string
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TenantID  uuid.UUID
	TokenHash string
	DeviceID  *string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (t RefreshToken) IsValid() bool {
	return t.RevokedAt == nil && t.ExpiresAt.After(time.Now().UTC())
}

type UserDevice struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	DeviceID  string
	Platform  string
	PushToken *string
	AppVersion *string
}
