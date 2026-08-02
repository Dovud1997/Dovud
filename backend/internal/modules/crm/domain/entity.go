package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	BranchID      *uuid.UUID
	Code          string
	Name          string
	Type          string
	Inn           *string
	Status        string
	CreditLimit   float64
	BalanceCached float64
	Lat           *float64
	Lng           *float64
	Address       *string
	Version       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CustomerContact struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	FullName   string
	Phone      string
	Email      *string
	Position   *string
	IsPrimary  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CustomerAddress struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Label      string
	Address    string
	Lat        *float64
	Lng        *float64
	IsPrimary  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CustomerCategory struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
