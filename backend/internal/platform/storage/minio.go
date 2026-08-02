package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStore struct {
	client *minio.Client
	bucket string
}

func NewMinioStore(cfg config.MinioConfig) (*MinioStore, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	bucket := cfg.Bucket
	if bucket == "" {
		bucket = "sfa"
	}
	return &MinioStore{client: client, bucket: bucket}, nil
}

func (s *MinioStore) Driver() string { return "minio" }
func (s *MinioStore) Bucket() string { return s.bucket }

func (s *MinioStore) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}

func (s *MinioStore) PresignPut(ctx context.Context, objectKey, contentType string, ttl time.Duration) (string, error) {
	_ = contentType
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	u, err := s.client.PresignedPutObject(ctx, s.bucket, objectKey, ttl)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *MinioStore) PresignGet(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *MinioStore) Put(ctx context.Context, objectKey, contentType string, body io.Reader, size int64) error {
	opts := minio.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, body, size, opts)
	return err
}

func (s *MinioStore) Get(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *MinioStore) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}
