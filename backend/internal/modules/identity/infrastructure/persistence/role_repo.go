package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepo struct{ db *gorm.DB }

func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{db: db} }

func toRole(m RoleModel) *domain.Role {
	return &domain.Role{
		ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name,
		IsSystem: m.IsSystem, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *RoleRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Role, error) {
	var rows []RoleModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? OR tenant_id IS NULL", tenantID).
		Order("code").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Role, 0, len(rows))
	for _, m := range rows {
		out = append(out, *toRole(m))
	}
	return out, nil
}

func (r *RoleRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Role, error) {
	var m RoleModel
	err := r.db.WithContext(ctx).Where("id = ? AND (tenant_id = ? OR tenant_id IS NULL)", id, tenantID).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toRole(m), nil
}

func (r *RoleRepo) FindByCode(ctx context.Context, tenantID *uuid.UUID, code string) (*domain.Role, error) {
	var m RoleModel
	q := r.db.WithContext(ctx).Where("code = ?", code)
	if tenantID == nil {
		q = q.Where("tenant_id IS NULL")
	} else {
		q = q.Where("tenant_id = ? OR tenant_id IS NULL", *tenantID)
	}
	err := q.First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toRole(m), nil
}

func (r *RoleRepo) Create(ctx context.Context, role *domain.Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	now := time.Now().UTC()
	role.CreatedAt, role.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&RoleModel{
		ID: role.ID, TenantID: role.TenantID, Code: role.Code, Name: role.Name,
		IsSystem: role.IsSystem, CreatedAt: role.CreatedAt, UpdatedAt: role.UpdatedAt,
	}).Error
}

func (r *RoleRepo) Update(ctx context.Context, role *domain.Role) error {
	role.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&RoleModel{}).Where("id = ?", role.ID).Updates(map[string]any{
		"name": role.Name, "updated_at": role.UpdatedAt,
	}).Error
}

func (r *RoleRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND is_system = ?", id, tenantID, false).Delete(&RoleModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *RoleRepo) SetPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RolePermissionModel{}).Error; err != nil {
			return err
		}
		for _, pid := range permissionIDs {
			if err := tx.Create(&RolePermissionModel{RoleID: roleID, PermissionID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RoleRepo) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	var rows []PermissionModel
	if err := r.db.WithContext(ctx).Order("code").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Permission, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Permission{ID: m.ID, Code: m.Code, Description: m.Description})
	}
	return out, nil
}

func (r *RoleRepo) PermissionIDsByCodes(ctx context.Context, codes []string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&PermissionModel{}).Where("code IN ?", codes).Pluck("id", &ids).Error
	return ids, err
}
