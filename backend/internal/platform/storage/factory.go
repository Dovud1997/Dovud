package storage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
)

func Open(cfg config.MinioConfig, publicBaseURL, signSecret string, log *slog.Logger) (ObjectStore, error) {
	if log == nil {
		log = slog.Default()
	}
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		driver = "auto"
	}

	switch driver {
	case "local":
		return NewLocalStore(cfg.LocalPath, cfg.Bucket, publicBaseURL, signSecret)
	case "minio":
		store, err := NewMinioStore(cfg)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := store.EnsureBucket(ctx); err != nil {
			return nil, fmt.Errorf("minio ensure bucket: %w", err)
		}
		return store, nil
	case "auto":
		store, err := NewMinioStore(cfg)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err = store.EnsureBucket(ctx)
			cancel()
		}
		if err == nil {
			log.Info("object storage ready", "driver", "minio", "bucket", store.Bucket())
			return store, nil
		}
		log.Warn("minio unavailable, using local storage", "error", err)
		local, lerr := NewLocalStore(cfg.LocalPath, cfg.Bucket, publicBaseURL, signSecret)
		if lerr != nil {
			return nil, lerr
		}
		return local, nil
	default:
		return nil, fmt.Errorf("unknown storage driver %q", driver)
	}
}
