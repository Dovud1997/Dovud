package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	tenantdomain "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type AuthService struct {
	users     domain.UserRepository
	tokens    domain.RefreshTokenRepository
	devices   domain.DeviceRepository
	tenants   tenantdomain.TenantRepository
	tokenSvc  *auth.TokenService
	lockout   auth.LoginGuard
}

func NewAuthService(
	users domain.UserRepository,
	tokens domain.RefreshTokenRepository,
	devices domain.DeviceRepository,
	tenants tenantdomain.TenantRepository,
	tokenSvc *auth.TokenService,
) *AuthService {
	return &AuthService{users: users, tokens: tokens, devices: devices, tenants: tenants, tokenSvc: tokenSvc}
}

func (s *AuthService) WithLoginGuard(g auth.LoginGuard) *AuthService {
	s.lockout = g
	return s
}

type LoginInput struct {
	TenantCode string `json:"tenant_code"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceID   string `json:"device_id"`
	Platform   string `json:"platform"`
}

type UserDTO struct {
	ID              uuid.UUID   `json:"id"`
	TenantID        uuid.UUID   `json:"tenant_id"`
	Email           string      `json:"email"`
	FullName        string      `json:"full_name"`
	Phone           *string     `json:"phone,omitempty"`
	Locale          string      `json:"locale"`
	ThemePreference string      `json:"theme_preference"`
	Status          string      `json:"status"`
	Roles           []string    `json:"roles"`
	RoleIDs         []uuid.UUID `json:"role_ids,omitempty"`
	Permissions     []string    `json:"permissions"`
	IsPlatformAdmin bool        `json:"is_platform_admin"`
}

type AuthResult struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	ExpiresIn        int64     `json:"expires_in"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	User             UserDTO   `json:"user"`
}

func hashRefresh(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.TenantCode = strings.TrimSpace(strings.ToLower(in.TenantCode))
	if in.Email == "" || in.Password == "" || in.TenantCode == "" {
		return nil, apperrors.ErrValidation
	}

	lockKey := in.TenantCode + ":" + in.Email
	if s.lockout != nil {
		if err := s.lockout.Check(ctx, lockKey); err != nil {
			return nil, err
		}
	}

	fail := func() (*AuthResult, error) {
		if s.lockout != nil {
			_ = s.lockout.Fail(ctx, lockKey)
		}
		return nil, apperrors.ErrInvalidCreds
	}

	tenant, err := s.tenants.FindByCode(ctx, in.TenantCode)
	if err != nil {
		return fail()
	}
	if tenant.Status != "active" {
		return nil, apperrors.New("TENANT_INACTIVE", "Tenant is inactive", 403)
	}

	user, err := s.users.FindByEmail(ctx, tenant.ID, in.Email)
	if err != nil {
		return fail()
	}
	if !user.IsActive() {
		return fail()
	}

	ok, err := auth.VerifyPassword(user.PasswordHash, in.Password)
	if err != nil || !ok {
		return fail()
	}
	if s.lockout != nil {
		_ = s.lockout.Success(ctx, lockKey)
	}

	return s.issueSession(ctx, user, in.DeviceID, in.Platform)
}

func (s *AuthService) issueSession(ctx context.Context, user *domain.User, deviceID, platform string) (*AuthResult, error) {
	roles, err := s.users.GetRoleCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	perms, err := s.users.GetPermissionCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	access, _, _, err := s.tokenSvc.IssueAccessToken(user.ID, user.TenantID, roles, perms, deviceID, user.IsPlatformAdmin)
	if err != nil {
		return nil, err
	}
	refreshPlain, refreshExp, err := s.tokenSvc.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	var devicePtr *string
	if deviceID != "" {
		devicePtr = &deviceID
	}
	if err := s.tokens.Create(ctx, &domain.RefreshToken{
		UserID: user.ID, TenantID: user.TenantID, TokenHash: hashRefresh(refreshPlain),
		DeviceID: devicePtr, ExpiresAt: refreshExp,
	}); err != nil {
		return nil, err
	}

	if deviceID != "" {
		_ = s.devices.Upsert(ctx, &domain.UserDevice{
			TenantID: user.TenantID, UserID: user.ID, DeviceID: deviceID,
			Platform: firstNonEmpty(platform, "unknown"),
		})
	}

	now := time.Now().UTC()
	user.LastLoginAt = &now
	_ = s.users.Update(ctx, user)

	return &AuthResult{
		AccessToken: access, RefreshToken: refreshPlain, TokenType: "Bearer",
		ExpiresIn: s.tokenSvc.AccessTTLSeconds(), RefreshExpiresAt: refreshExp,
		User: toUserDTO(user, roles, perms),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken, deviceID string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, apperrors.ErrTokenInvalid
	}
	stored, err := s.tokens.FindByHash(ctx, hashRefresh(refreshToken))
	if err != nil || !stored.IsValid() {
		return nil, apperrors.ErrTokenInvalid
	}
	_ = s.tokens.Revoke(ctx, stored.ID)

	user, err := s.users.FindByIDAnyTenant(ctx, stored.UserID)
	if err != nil || !user.IsActive() {
		return nil, apperrors.ErrTokenInvalid
	}
	if deviceID == "" && stored.DeviceID != nil {
		deviceID = *stored.DeviceID
	}
	return s.issueSession(ctx, user, deviceID, "")
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	stored, err := s.tokens.FindByHash(ctx, hashRefresh(refreshToken))
	if err != nil {
		return nil
	}
	return s.tokens.Revoke(ctx, stored.ID)
}

func (s *AuthService) Me(ctx context.Context, tenantID, userID uuid.UUID) (*UserDTO, error) {
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	roles, err := s.users.GetRoleCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	perms, err := s.users.GetPermissionCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	dto := toUserDTO(user, roles, perms)
	return &dto, nil
}

type UpdateMeInput struct {
	FullName        *string `json:"full_name"`
	Locale          *string `json:"locale"`
	ThemePreference *string `json:"theme_preference"`
	Phone           *string `json:"phone"`
}

func (s *AuthService) UpdateMe(ctx context.Context, tenantID, userID uuid.UUID, in UpdateMeInput) (*UserDTO, error) {
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if in.FullName != nil {
		user.FullName = strings.TrimSpace(*in.FullName)
	}
	if in.Locale != nil {
		user.Locale = *in.Locale
	}
	if in.ThemePreference != nil {
		user.ThemePreference = *in.ThemePreference
	}
	if in.Phone != nil {
		user.Phone = in.Phone
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return s.Me(ctx, tenantID, userID)
}

func (s *AuthService) ChangePassword(ctx context.Context, tenantID, userID uuid.UUID, current, next string) error {
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	ok, err := auth.VerifyPassword(user.PasswordHash, current)
	if err != nil || !ok {
		return apperrors.ErrInvalidCreds
	}
	if len(next) < 8 {
		return apperrors.ErrValidation
	}
	hash, err := auth.HashPassword(next)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}
	return s.tokens.RevokeAllForUser(ctx, user.ID)
}

type RegisterDeviceInput struct {
	DeviceID   string  `json:"device_id"`
	Platform   string  `json:"platform"`
	PushToken  *string `json:"push_token"`
	AppVersion *string `json:"app_version"`
}

type DeviceDTO struct {
	ID         uuid.UUID `json:"id"`
	DeviceID   string    `json:"device_id"`
	Platform   string    `json:"platform"`
	PushToken  *string   `json:"push_token,omitempty"`
	AppVersion *string   `json:"app_version,omitempty"`
}

func (s *AuthService) RegisterDevice(ctx context.Context, tenantID, userID uuid.UUID, in RegisterDeviceInput) (*DeviceDTO, error) {
	deviceKey := strings.TrimSpace(in.DeviceID)
	platform := strings.TrimSpace(in.Platform)
	if deviceKey == "" || platform == "" {
		return nil, apperrors.ErrValidation
	}
	dev := &domain.UserDevice{
		TenantID: tenantID, UserID: userID, DeviceID: deviceKey,
		Platform: platform, PushToken: in.PushToken, AppVersion: in.AppVersion,
	}
	if err := s.devices.Upsert(ctx, dev); err != nil {
		return nil, err
	}
	rows, err := s.devices.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	for _, d := range rows {
		if d.DeviceID == deviceKey {
			return &DeviceDTO{
				ID: d.ID, DeviceID: d.DeviceID, Platform: d.Platform,
				PushToken: d.PushToken, AppVersion: d.AppVersion,
			}, nil
		}
	}
	return &DeviceDTO{DeviceID: deviceKey, Platform: platform, PushToken: in.PushToken, AppVersion: in.AppVersion}, nil
}

func (s *AuthService) UnregisterDevice(ctx context.Context, userID uuid.UUID, deviceKey string) error {
	deviceKey = strings.TrimSpace(deviceKey)
	if deviceKey == "" {
		return apperrors.ErrValidation
	}
	return s.devices.DeleteByDeviceKey(ctx, userID, deviceKey)
}

func (s *AuthService) ListDevices(ctx context.Context, tenantID, userID uuid.UUID) ([]DeviceDTO, error) {
	rows, err := s.devices.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]DeviceDTO, 0, len(rows))
	for _, d := range rows {
		out = append(out, DeviceDTO{
			ID: d.ID, DeviceID: d.DeviceID, Platform: d.Platform,
			PushToken: d.PushToken, AppVersion: d.AppVersion,
		})
	}
	return out, nil
}

func toUserDTO(u *domain.User, roles, perms []string) UserDTO {
	if roles == nil {
		roles = []string{}
	}
	if perms == nil {
		perms = []string{}
	}
	return UserDTO{
		ID: u.ID, TenantID: u.TenantID, Email: u.Email, FullName: u.FullName, Phone: u.Phone,
		Locale: u.Locale, ThemePreference: u.ThemePreference, Status: u.Status,
		Roles: roles, RoleIDs: []uuid.UUID{}, Permissions: perms, IsPlatformAdmin: u.IsPlatformAdmin,
	}
}

func firstNonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
