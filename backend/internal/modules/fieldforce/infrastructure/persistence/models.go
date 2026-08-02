package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesAgentModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	BranchID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	EmployeeCode string     `gorm:"size:64;not null"`
	ManagerID    *uuid.UUID `gorm:"type:uuid;index"`
	Status       string     `gorm:"size:32;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (SalesAgentModel) TableName() string { return "sales_agents" }

type RouteModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index"`
	AgentID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Date      time.Time `gorm:"type:date;not null"`
	Name      string    `gorm:"size:255;not null"`
	Status    string    `gorm:"size:32;not null"`
	Version   int64     `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (RouteModel) TableName() string { return "routes" }

type RouteStopModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	RouteID        uuid.UUID `gorm:"type:uuid;not null;index"`
	CustomerID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Sequence       int       `gorm:"not null"`
	PlannedArrival *time.Time
	Status         string `gorm:"size:32;not null"`
}

func (RouteStopModel) TableName() string { return "route_stops" }

type VisitModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	AgentID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	CustomerID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	RouteStopID *uuid.UUID `gorm:"type:uuid;index"`
	StartedAt   time.Time  `gorm:"not null"`
	EndedAt     *time.Time
	CheckinLat  *float64
	CheckinLng  *float64
	CheckoutLat *float64
	CheckoutLng *float64
	Result      string `gorm:"size:32;not null;default:''"`
	Notes       *string
	Version     int64 `gorm:"not null;default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (VisitModel) TableName() string { return "visits" }

type VisitPhotoModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	VisitID   uuid.UUID `gorm:"type:uuid;not null;index"`
	FileURL   string    `gorm:"not null"`
	Caption   *string
	CreatedAt time.Time
}

func (VisitPhotoModel) TableName() string { return "visit_photos" }

type VisitCommentModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	VisitID      uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorUserID uuid.UUID `gorm:"type:uuid;not null"`
	Body         string    `gorm:"not null"`
	CreatedAt    time.Time
}

func (VisitCommentModel) TableName() string { return "visit_comments" }

type GpsTrackModel struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	AgentID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	VisitID    *uuid.UUID `gorm:"type:uuid;index"`
	Lat        float64    `gorm:"not null"`
	Lng        float64    `gorm:"not null"`
	Accuracy   *float64
	RecordedAt time.Time `gorm:"not null;index"`
	CreatedAt  time.Time
}

func (GpsTrackModel) TableName() string { return "gps_tracks" }
