package domain

import (
	"context"

	"github.com/google/uuid"
)

type FileRepository interface {
	Create(ctx context.Context, f *File) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*File, error)
	Update(ctx context.Context, f *File) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]File, int64, error)
}

type DocumentRepository interface {
	Create(ctx context.Context, d *Document) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Document, error)
	Update(ctx context.Context, d *Document) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]Document, int64, error)
	ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, page, perPage int) ([]Document, int64, error)
	AttachFile(ctx context.Context, documentID, fileID uuid.UUID, role string) error
	ListFiles(ctx context.Context, documentID uuid.UUID) ([]File, error)
}
