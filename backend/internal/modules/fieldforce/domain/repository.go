package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AgentRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]SalesAgent, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*SalesAgent, error)
	FindByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*SalesAgent, error)
	Create(ctx context.Context, a *SalesAgent) error
	Update(ctx context.Context, a *SalesAgent) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}

type RouteRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, agentID *uuid.UUID, date *time.Time, page, perPage int) ([]Route, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Route, error)
	Create(ctx context.Context, r *Route) error
	Update(ctx context.Context, r *Route) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
	ReplaceStops(ctx context.Context, routeID uuid.UUID, stops []RouteStop) error
	ListStops(ctx context.Context, routeID uuid.UUID) ([]RouteStop, error)
}

type VisitRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, agentID, customerID *uuid.UUID, page, perPage int) ([]Visit, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Visit, error)
	Create(ctx context.Context, v *Visit) error
	Update(ctx context.Context, v *Visit) error
	AddPhoto(ctx context.Context, p *VisitPhoto) error
	ListPhotos(ctx context.Context, visitID uuid.UUID) ([]VisitPhoto, error)
	AddComment(ctx context.Context, c *VisitComment) error
	ListComments(ctx context.Context, visitID uuid.UUID) ([]VisitComment, error)
}

type GpsRepository interface {
	AddBatch(ctx context.Context, points []GpsPoint) error
	ListByAgent(ctx context.Context, tenantID, agentID uuid.UUID, from, to time.Time) ([]GpsPoint, error)
	LatestByAgent(ctx context.Context, tenantID, agentID uuid.UUID) (*GpsPoint, error)
}
