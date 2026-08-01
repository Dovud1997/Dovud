package application

import (
	"context"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/tenant/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type TenantService struct {
	tenants  domain.TenantRepository
	branding domain.BrandingRepository
	domains  domain.DomainRepository
}

func NewTenantService(
	tenants domain.TenantRepository,
	branding domain.BrandingRepository,
	domains domain.DomainRepository,
) *TenantService {
	return &TenantService{tenants: tenants, branding: branding, domains: domains}
}

type TenantDTO struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	DefaultLocale   string    `json:"default_locale"`
	DefaultCurrency string    `json:"default_currency"`
	Timezone        string    `json:"timezone"`
}

func (s *TenantService) Get(ctx context.Context, tenantID uuid.UUID) (*TenantDTO, error) {
	t, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &TenantDTO{
		ID: t.ID, Code: t.Code, Name: t.Name, Status: t.Status,
		DefaultLocale: t.DefaultLocale, DefaultCurrency: t.DefaultCurrency, Timezone: t.Timezone,
	}, nil
}

type UpdateTenantInput struct {
	Name            *string `json:"name"`
	DefaultLocale   *string `json:"default_locale"`
	DefaultCurrency *string `json:"default_currency"`
	Timezone        *string `json:"timezone"`
}

func (s *TenantService) Update(ctx context.Context, tenantID uuid.UUID, in UpdateTenantInput) (*TenantDTO, error) {
	t, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		t.Name = strings.TrimSpace(*in.Name)
	}
	if in.DefaultLocale != nil {
		t.DefaultLocale = *in.DefaultLocale
	}
	if in.DefaultCurrency != nil {
		t.DefaultCurrency = strings.ToUpper(*in.DefaultCurrency)
	}
	if in.Timezone != nil {
		t.Timezone = *in.Timezone
	}
	if err := s.tenants.Update(ctx, t); err != nil {
		return nil, err
	}
	return s.Get(ctx, tenantID)
}

type BrandingDTO struct {
	AppName          string  `json:"app_name"`
	LogoURL          *string `json:"logo_url"`
	FaviconURL       *string `json:"favicon_url"`
	IconURL          *string `json:"icon_url"`
	PrimaryColor     string  `json:"primary_color"`
	SecondaryColor   string  `json:"secondary_color"`
	AccentColor      string  `json:"accent_color"`
	ThemeModeDefault string  `json:"theme_mode_default"`
	BrandingVersion  int64   `json:"branding_version"`
}

func (s *TenantService) GetBranding(ctx context.Context, tenantID uuid.UUID) (*BrandingDTO, error) {
	b, err := s.branding.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &BrandingDTO{
		AppName: b.AppName, LogoURL: b.LogoURL, FaviconURL: b.FaviconURL, IconURL: b.IconURL,
		PrimaryColor: b.PrimaryColor, SecondaryColor: b.SecondaryColor, AccentColor: b.AccentColor,
		ThemeModeDefault: b.ThemeModeDefault, BrandingVersion: b.BrandingVersion,
	}, nil
}

type UpdateBrandingInput struct {
	AppName          *string `json:"app_name"`
	LogoURL          *string `json:"logo_url"`
	FaviconURL       *string `json:"favicon_url"`
	IconURL          *string `json:"icon_url"`
	PrimaryColor     *string `json:"primary_color"`
	SecondaryColor   *string `json:"secondary_color"`
	AccentColor      *string `json:"accent_color"`
	ThemeModeDefault *string `json:"theme_mode_default"`
}

func (s *TenantService) UpdateBranding(ctx context.Context, tenantID uuid.UUID, in UpdateBrandingInput) (*BrandingDTO, error) {
	b, err := s.branding.GetByTenantID(ctx, tenantID)
	if err != nil {
		b = &domain.Branding{
			TenantID: tenantID, AppName: "SFA",
			PrimaryColor: "#0F766E", SecondaryColor: "#134E4A", AccentColor: "#F59E0B",
			ThemeModeDefault: "light",
		}
	}
	if in.AppName != nil {
		b.AppName = strings.TrimSpace(*in.AppName)
	}
	if in.LogoURL != nil {
		b.LogoURL = in.LogoURL
	}
	if in.FaviconURL != nil {
		b.FaviconURL = in.FaviconURL
	}
	if in.IconURL != nil {
		b.IconURL = in.IconURL
	}
	if in.PrimaryColor != nil {
		b.PrimaryColor = *in.PrimaryColor
	}
	if in.SecondaryColor != nil {
		b.SecondaryColor = *in.SecondaryColor
	}
	if in.AccentColor != nil {
		b.AccentColor = *in.AccentColor
	}
	if in.ThemeModeDefault != nil {
		b.ThemeModeDefault = *in.ThemeModeDefault
	}
	if strings.TrimSpace(b.AppName) == "" {
		return nil, apperrors.ErrValidation
	}
	if err := s.branding.Upsert(ctx, b); err != nil {
		return nil, err
	}
	return s.GetBranding(ctx, tenantID)
}

func (s *TenantService) PublicBranding(ctx context.Context, code, host string) (*domain.PublicBranding, error) {
	if code != "" {
		return s.branding.GetPublicByCode(ctx, code)
	}
	if host != "" {
		return s.branding.GetPublicByHost(ctx, host)
	}
	return nil, apperrors.ErrValidation
}

type DomainDTO struct {
	ID        uuid.UUID `json:"id"`
	Host      string    `json:"host"`
	IsPrimary bool      `json:"is_primary"`
}

func (s *TenantService) ListDomains(ctx context.Context, tenantID uuid.UUID) ([]DomainDTO, error) {
	rows, err := s.domains.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]DomainDTO, 0, len(rows))
	for _, d := range rows {
		out = append(out, DomainDTO{ID: d.ID, Host: d.Host, IsPrimary: d.IsPrimary})
	}
	return out, nil
}

type CreateDomainInput struct {
	Host      string `json:"host"`
	IsPrimary bool   `json:"is_primary"`
}

func (s *TenantService) AddDomain(ctx context.Context, tenantID uuid.UUID, in CreateDomainInput) (*DomainDTO, error) {
	host := strings.ToLower(strings.TrimSpace(in.Host))
	if host == "" {
		return nil, apperrors.ErrValidation
	}
	d := &domain.Domain{TenantID: tenantID, Host: host, IsPrimary: in.IsPrimary}
	if err := s.domains.Create(ctx, d); err != nil {
		return nil, err
	}
	return &DomainDTO{ID: d.ID, Host: d.Host, IsPrimary: d.IsPrimary}, nil
}

func (s *TenantService) DeleteDomain(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.domains.Delete(ctx, tenantID, id)
}
