package persistence

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func toUser(m UserModel) *domain.User {
	var deleted *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deleted = &t
	}
	return &domain.User{
		ID: m.ID, TenantID: m.TenantID, Email: m.Email, Phone: m.Phone,
		PasswordHash: m.PasswordHash, FullName: m.FullName, Status: m.Status,
		Locale: m.Locale, ThemePreference: m.ThemePreference, LastLoginAt: m.LastLoginAt,
		IsPlatformAdmin: m.IsPlatformAdmin, Version: m.Version,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, DeletedAt: deleted,
	}
}

func (r *UserRepo) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND lower(email) = ?", tenantID, strings.ToLower(email)).
		First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUser(m), nil
}

func (r *UserRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUser(m), nil
}

func (r *UserRepo) FindByIDAnyTenant(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUser(m), nil
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	m := UserModel{
		ID: user.ID, TenantID: user.TenantID, Email: strings.ToLower(user.Email), Phone: user.Phone,
		PasswordHash: user.PasswordHash, FullName: user.FullName, Status: user.Status,
		Locale: user.Locale, ThemePreference: user.ThemePreference, IsPlatformAdmin: user.IsPlatformAdmin,
		Version: user.Version, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now().UTC()
	user.Version++
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ? AND tenant_id = ?", user.ID, user.TenantID).Updates(map[string]any{
		"email":            strings.ToLower(user.Email),
		"phone":            user.Phone,
		"password_hash":    user.PasswordHash,
		"full_name":        user.FullName,
		"status":           user.Status,
		"locale":           user.Locale,
		"theme_preference": user.ThemePreference,
		"last_login_at":    user.LastLoginAt,
		"version":          user.Version,
		"updated_at":       user.UpdatedAt,
	}).Error
}

func (r *UserRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	var total int64
	q := r.db.WithContext(ctx).Model(&UserModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []UserModel
	err := q.Order("created_at DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.User, 0, len(rows))
	for _, m := range rows {
		out = append(out, *toUser(m))
	}
	return out, total, nil
}

func (r *UserRepo) ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&UserRoleModel{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Create(&UserRoleModel{UserID: userID, RoleID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepo) GetRoleCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Pluck("roles.code", &codes).Error
	return codes, err
}

func (r *UserRepo) GetRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&UserRoleModel{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &ids).Error
	return ids, err
}

func (r *UserRepo) GetPermissionCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).Table("permissions").
		Select("DISTINCT permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Pluck("permissions.code", &codes).Error
	return codes, err
}
