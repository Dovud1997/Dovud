package domain

import (
	"context"

	"github.com/google/uuid"
)

type CustomerUserRepository interface {
	FindByUser(ctx context.Context, tenantID, userID uuid.UUID) (*CustomerUser, error)
	Upsert(ctx context.Context, link *CustomerUser) error
}
