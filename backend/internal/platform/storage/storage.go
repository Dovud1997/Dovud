package storage

import (
	"context"
	"io"
	"time"
)

// ObjectStore abstracts blob storage (MinIO/S3 or local filesystem).
type ObjectStore interface {
	Driver() string
	Bucket() string
	EnsureBucket(ctx context.Context) error
	PresignPut(ctx context.Context, objectKey, contentType string, ttl time.Duration) (string, error)
	PresignGet(ctx context.Context, objectKey string, ttl time.Duration) (string, error)
	Put(ctx context.Context, objectKey, contentType string, body io.Reader, size int64) error
	Get(ctx context.Context, objectKey string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectKey string) error
}
