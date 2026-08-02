package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantModel struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Code            string         `gorm:"size:64;not null"`
	Name            string         `gorm:"size:255;not null"`
	Status          string         `gorm:"size:32;not null"`
	DefaultLocale   string         `gorm:"size:8;not null"`
	DefaultCurrency string         `gorm:"size:3;not null"`
	Timezone        string         `gorm:"size:64;not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (TenantModel) TableName() string { return "tenants" }

type BrandingModel struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID         uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	AppName          string    `gorm:"size:255;not null"`
	LogoURL          *string
	FaviconURL       *string
	IconURL          *string
	PrimaryColor     string `gorm:"size:32;not null"`
	SecondaryColor   string `gorm:"size:32;not null"`
	AccentColor      string `gorm:"size:32;not null"`
	ThemeModeDefault string `gorm:"size:16;not null"`
	BrandingVersion  int64  `gorm:"not null;default:1"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (BrandingModel) TableName() string { return "tenant_branding" }

type DomainModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	Host      string         `gorm:"size:255;not null"`
	IsPrimary bool           `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (DomainModel) TableName() string { return "tenant_domains" }

type ProviderModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_tenant_provider_type"`
	Type      string    `gorm:"size:32;not null;uniqueIndex:uq_tenant_provider_type"`
	Driver    string    `gorm:"size:32;not null;default:log"`
	ConfigEnc string    `gorm:"type:text;not null"`
	IsEnabled bool      `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ProviderModel) TableName() string { return "tenant_providers" }
