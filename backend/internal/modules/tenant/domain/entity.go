package domain

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID              uuid.UUID
	Code            string
	Name            string
	Status          string
	DefaultLocale   string
	DefaultCurrency string
	Timezone        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Branding struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	AppName           string
	LogoURL           *string
	FaviconURL        *string
	IconURL           *string
	PrimaryColor      string
	SecondaryColor    string
	AccentColor       string
	ThemeModeDefault  string
	BrandingVersion   int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Domain struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Host      string
	IsPrimary bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	ProviderSMTP = "smtp"
	ProviderSMS  = "sms"
	ProviderPush = "push"
)

type TenantProvider struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Type      string
	Driver    string
	ConfigEnc string
	IsEnabled bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ValidProviderType(t string) bool {
	switch t {
	case ProviderSMTP, ProviderSMS, ProviderPush:
		return true
	default:
		return false
	}
}

type PublicBranding struct {
	TenantCode       string  `json:"tenant_code"`
	TenantName       string  `json:"tenant_name"`
	AppName          string  `json:"app_name"`
	LogoURL          *string `json:"logo_url"`
	FaviconURL       *string `json:"favicon_url"`
	IconURL          *string `json:"icon_url"`
	PrimaryColor     string  `json:"primary_color"`
	SecondaryColor   string  `json:"secondary_color"`
	AccentColor      string  `json:"accent_color"`
	ThemeModeDefault string  `json:"theme_mode_default"`
	BrandingVersion  int64   `json:"branding_version"`
	DefaultLocale    string  `json:"default_locale"`
}
