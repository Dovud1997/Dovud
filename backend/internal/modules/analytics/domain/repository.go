package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SnapshotFilters struct {
	Code   string
	Period string
}

type KpiRepository interface {
	ListDefinitions(ctx context.Context, tenantID uuid.UUID) ([]KpiDefinition, error)
	CreateDefinition(ctx context.Context, d *KpiDefinition) error
	ListSnapshots(ctx context.Context, tenantID uuid.UUID, filters SnapshotFilters, page, perPage int) ([]KpiSnapshot, int64, error)
	CreateSnapshot(ctx context.Context, s *KpiSnapshot) error
}

type SalesPoint struct {
	Date   time.Time `json:"date"`
	Orders int64     `json:"orders"`
	Amount float64   `json:"amount"`
}

type VisitPoint struct {
	Date   time.Time `json:"date"`
	Visits int64     `json:"visits"`
}
