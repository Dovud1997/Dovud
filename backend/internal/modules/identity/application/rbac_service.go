package application

import (
	"context"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type RBACService struct {
	users domain.UserRepository
	roles domain.RoleRepository
}

func NewRBACService(users domain.UserRepository, roles domain.RoleRepository) *RBACService {
	return &RBACService{users: users, roles: roles}
}

type CreateUserInput struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	FullName string      `json:"full_name"`
	Phone    *string     `json:"phone"`
	Locale   string      `json:"locale"`
	RoleIDs  []uuid.UUID `json:"role_ids"`
}

func (s *RBACService) ListUsers(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]UserDTO, int64, error) {
	users, total, err := s.users.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]UserDTO, 0, len(users))
	for i := range users {
		roles, _ := s.users.GetRoleCodes(ctx, users[i].ID)
		perms, _ := s.users.GetPermissionCodes(ctx, users[i].ID)
		roleIDs, _ := s.users.GetRoleIDs(ctx, users[i].ID)
		dto := toUserDTO(&users[i], roles, perms)
		if roleIDs == nil {
			roleIDs = []uuid.UUID{}
		}
		dto.RoleIDs = roleIDs
		out = append(out, dto)
	}
	return out, total, nil
}

func (s *RBACService) CreateUser(ctx context.Context, tenantID uuid.UUID, in CreateUserInput) (*UserDTO, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" || strings.TrimSpace(in.FullName) == "" {
		return nil, apperrors.ErrValidation
	}
	if _, err := s.users.FindByEmail(ctx, tenantID, email); err == nil {
		return nil, apperrors.New("USER_EXISTS", "User already exists", 409)
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	locale := in.Locale
	if locale == "" {
		locale = "ru"
	}
	user := &domain.User{
		TenantID: tenantID, Email: email, Phone: in.Phone, PasswordHash: hash,
		FullName: strings.TrimSpace(in.FullName), Status: "active", Locale: locale,
		ThemePreference: "system", Version: 1,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	if len(in.RoleIDs) > 0 {
		if err := s.users.ReplaceRoles(ctx, user.ID, in.RoleIDs); err != nil {
			return nil, err
		}
	}
	roles, _ := s.users.GetRoleCodes(ctx, user.ID)
	perms, _ := s.users.GetPermissionCodes(ctx, user.ID)
	roleIDs, _ := s.users.GetRoleIDs(ctx, user.ID)
	dto := toUserDTO(user, roles, perms)
	if roleIDs == nil {
		roleIDs = []uuid.UUID{}
	}
	dto.RoleIDs = roleIDs
	return &dto, nil
}

type UpdateUserInput struct {
	FullName *string `json:"full_name"`
	Phone    *string `json:"phone"`
	Locale   *string `json:"locale"`
	Status   *string `json:"status"`
}

func (s *RBACService) UpdateUser(ctx context.Context, tenantID, userID uuid.UUID, in UpdateUserInput) (*UserDTO, error) {
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if in.FullName != nil {
		name := strings.TrimSpace(*in.FullName)
		if name == "" {
			return nil, apperrors.ErrValidation
		}
		user.FullName = name
	}
	if in.Phone != nil {
		phone := strings.TrimSpace(*in.Phone)
		if phone == "" {
			user.Phone = nil
		} else {
			user.Phone = &phone
		}
	}
	if in.Locale != nil {
		user.Locale = strings.TrimSpace(*in.Locale)
	}
	if in.Status != nil {
		st := strings.TrimSpace(*in.Status)
		switch st {
		case "active", "disabled", "invited":
			user.Status = st
		default:
			return nil, apperrors.ErrValidation
		}
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	roles, _ := s.users.GetRoleCodes(ctx, user.ID)
	perms, _ := s.users.GetPermissionCodes(ctx, user.ID)
	roleIDs, _ := s.users.GetRoleIDs(ctx, user.ID)
	dto := toUserDTO(user, roles, perms)
	if roleIDs == nil {
		roleIDs = []uuid.UUID{}
	}
	dto.RoleIDs = roleIDs
	return &dto, nil
}

func (s *RBACService) AssignRoles(ctx context.Context, tenantID, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if _, err := s.users.FindByID(ctx, tenantID, userID); err != nil {
		return err
	}
	for _, rid := range roleIDs {
		if _, err := s.roles.FindByID(ctx, tenantID, rid); err != nil {
			return apperrors.ErrValidation
		}
	}
	return s.users.ReplaceRoles(ctx, userID, roleIDs)
}

type RoleDTO struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        *uuid.UUID `json:"tenant_id,omitempty"`
	Code            string     `json:"code"`
	Name            string     `json:"name"`
	IsSystem        bool       `json:"is_system"`
	PermissionCodes []string   `json:"permission_codes"`
}

func (s *RBACService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]RoleDTO, error) {
	roles, err := s.roles.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]RoleDTO, 0, len(roles))
	for _, r := range roles {
		codes, err := s.roles.PermissionCodesByRoleID(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		if codes == nil {
			codes = []string{}
		}
		out = append(out, RoleDTO{
			ID: r.ID, TenantID: r.TenantID, Code: r.Code, Name: r.Name,
			IsSystem: r.IsSystem, PermissionCodes: codes,
		})
	}
	return out, nil
}

type CreateRoleInput struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	PermissionCodes []string `json:"permission_codes"`
}

func (s *RBACService) CreateRole(ctx context.Context, tenantID uuid.UUID, in CreateRoleInput) (*RoleDTO, error) {
	code := strings.TrimSpace(strings.ToLower(in.Code))
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return nil, apperrors.ErrValidation
	}
	tid := tenantID
	role := &domain.Role{TenantID: &tid, Code: code, Name: name, IsSystem: false}
	if err := s.roles.Create(ctx, role); err != nil {
		return nil, err
	}
	if len(in.PermissionCodes) > 0 {
		ids, err := s.roles.PermissionIDsByCodes(ctx, in.PermissionCodes)
		if err != nil {
			return nil, err
		}
		if err := s.roles.SetPermissions(ctx, role.ID, ids); err != nil {
			return nil, err
		}
	}
	codes, _ := s.roles.PermissionCodesByRoleID(ctx, role.ID)
	if codes == nil {
		codes = []string{}
	}
	dto := RoleDTO{
		ID: role.ID, TenantID: role.TenantID, Code: role.Code, Name: role.Name,
		IsSystem: role.IsSystem, PermissionCodes: codes,
	}
	return &dto, nil
}

func (s *RBACService) SetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID, codes []string) error {
	role, err := s.roles.FindByID(ctx, tenantID, roleID)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return apperrors.New("ROLE_LOCKED", "System roles are immutable", 403)
	}
	ids, err := s.roles.PermissionIDsByCodes(ctx, codes)
	if err != nil {
		return err
	}
	return s.roles.SetPermissions(ctx, roleID, ids)
}

type PermissionDTO struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
}

func (s *RBACService) ListPermissions(ctx context.Context) ([]PermissionDTO, error) {
	perms, err := s.roles.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PermissionDTO, 0, len(perms))
	for _, p := range perms {
		out = append(out, PermissionDTO{ID: p.ID, Code: p.Code, Description: p.Description})
	}
	return out, nil
}
