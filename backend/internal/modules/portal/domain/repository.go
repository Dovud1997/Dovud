package domain

import (
	"context"

	"github.com/google/uuid"
)

type CustomerUserRepository interface {
	FindByUser(ctx context.Context, tenantID, userID uuid.UUID) (*CustomerUser, error)
	Upsert(ctx context.Context, link *CustomerUser) error
	DeleteByUser(ctx context.Context, tenantID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]CustomerUser, error)
}
