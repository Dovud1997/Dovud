package domain

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Name      string
	Inn       *string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Branch struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	CompanyID *uuid.UUID
	Code      string
	Name      string
	Address   *string
	Lat       *float64
	Lng       *float64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Warehouse struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	Code      string
	Name      string
	Type      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WarehouseStock struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	ProductID   uuid.UUID
	QtyOnHand   float64
	QtyReserved float64
	UpdatedAt   time.Time
}
