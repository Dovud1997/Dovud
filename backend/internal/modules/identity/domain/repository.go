package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*User, error)
	FindByIDAnyTenant(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]User, int64, error)
	ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	GetRoleCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetPermissionCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type RoleRepository interface {
	List(ctx context.Context, tenantID uuid.UUID) ([]Role, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Role, error)
	FindByCode(ctx context.Context, tenantID *uuid.UUID, code string) (*Role, error)
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
	SetPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	ListPermissions(ctx context.Context) ([]Permission, error)
	PermissionIDsByCodes(ctx context.Context, codes []string) ([]uuid.UUID, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type DeviceRepository interface {
	Upsert(ctx context.Context, device *UserDevice) error
	Delete(ctx context.Context, userID, deviceID uuid.UUID) error
}
