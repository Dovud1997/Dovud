package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	BranchID      *uuid.UUID     `gorm:"type:uuid;index"`
	Code          string         `gorm:"size:64;not null"`
	Name          string         `gorm:"size:255;not null"`
	Type          string         `gorm:"size:32;not null"`
	Inn           *string        `gorm:"size:64"`
	Status        string         `gorm:"size:32;not null"`
	CreditLimit   float64        `gorm:"not null;default:0"`
	BalanceCached float64        `gorm:"not null;default:0"`
	Lat           *float64
	Lng           *float64
	Address       *string
	Version       int64          `gorm:"not null;default:1"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (CustomerModel) TableName() string { return "customers" }

type CustomerContactModel struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CustomerID uuid.UUID      `gorm:"type:uuid;not null;index"`
	FullName   string         `gorm:"size:255;not null"`
	Phone      string         `gorm:"size:64;not null"`
	Email      *string        `gorm:"size:255"`
	Position   *string        `gorm:"size:128"`
	IsPrimary  bool           `gorm:"not null;default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (CustomerContactModel) TableName() string { return "customer_contacts" }

type CustomerAddressModel struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CustomerID uuid.UUID      `gorm:"type:uuid;not null;index"`
	Label      string         `gorm:"size:128;not null"`
	Address    string         `gorm:"not null"`
	Lat        *float64
	Lng        *float64
	IsPrimary  bool           `gorm:"not null;default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (CustomerAddressModel) TableName() string { return "customer_addresses" }

type CustomerCategoryModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	Code      string         `gorm:"size:64;not null"`
	Name      string         `gorm:"size:255;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CustomerCategoryModel) TableName() string { return "customer_categories" }
