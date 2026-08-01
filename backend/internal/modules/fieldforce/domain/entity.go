package domain

import (
	"time"

	"github.com/google/uuid"
)

type SalesAgent struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	UserID       uuid.UUID
	BranchID     uuid.UUID
	EmployeeCode string
	ManagerID    *uuid.UUID
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Route struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	AgentID   uuid.UUID
	Date      time.Time
	Name      string
	Status    string
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RouteStop struct {
	ID             uuid.UUID
	RouteID        uuid.UUID
	CustomerID     uuid.UUID
	Sequence       int
	PlannedArrival *time.Time
	Status         string
}

type Visit struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	AgentID     uuid.UUID
	CustomerID  uuid.UUID
	RouteStopID *uuid.UUID
	StartedAt   time.Time
	EndedAt     *time.Time
	CheckinLat  *float64
	CheckinLng  *float64
	CheckoutLat *float64
	CheckoutLng *float64
	Result      string
	Notes       *string
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type VisitPhoto struct {
	ID        uuid.UUID
	VisitID   uuid.UUID
	FileURL   string
	Caption   *string
	CreatedAt time.Time
}

type VisitComment struct {
	ID           uuid.UUID
	VisitID      uuid.UUID
	AuthorUserID uuid.UUID
	Body         string
	CreatedAt    time.Time
}

type GpsPoint struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	AgentID    uuid.UUID
	VisitID    *uuid.UUID
	Lat        float64
	Lng        float64
	Accuracy   *float64
	RecordedAt time.Time
	CreatedAt  time.Time
}
