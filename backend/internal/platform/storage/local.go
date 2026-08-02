package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LocalStore stores objects under a directory. Presigned URLs point at the API
// helpers /api/v1/files/local/{put|get} with an HMAC signature.
type LocalStore struct {
	root          string
	bucket        string
	publicBaseURL string
	signSecret    string
}

func NewLocalStore(root, bucket, publicBaseURL, signSecret string) (*LocalStore, error) {
	if root == "" {
		root = "./storage"
	}
	if bucket == "" {
		bucket = "sfa"
	}
	if signSecret == "" {
		signSecret = "local-dev-secret"
	}
	dir := filepath.Join(root, bucket)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &LocalStore{
		root: root, bucket: bucket,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		signSecret:    signSecret,
	}, nil
}

func (s *LocalStore) Driver() string { return "local" }
func (s *LocalStore) Bucket() string { return s.bucket }

func (s *LocalStore) pathFor(objectKey string) (string, error) {
	clean := filepath.Clean("/" + objectKey)
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || strings.Contains(clean, "..") {
		return "", fmt.Errorf("invalid object key")
	}
	return filepath.Join(s.root, s.bucket, filepath.FromSlash(clean)), nil
}

func (s *LocalStore) EnsureBucket(ctx context.Context) error {
	_ = ctx
	return os.MkdirAll(filepath.Join(s.root, s.bucket), 0o755)
}

func (s *LocalStore) sign(op, objectKey string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(s.signSecret))
	_, _ = mac.Write([]byte(fmt.Sprintf("%s\n%s\n%d", op, objectKey, exp)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *LocalStore) Verify(op, objectKey, sig string, exp int64) bool {
	if exp < time.Now().Unix() {
		return false
	}
	expected := s.sign(op, objectKey, exp)
	return hmac.Equal([]byte(expected), []byte(sig))
}

func (s *LocalStore) signedURL(op, objectKey string, ttl time.Duration) string {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	exp := time.Now().Add(ttl).Unix()
	sig := s.sign(op, objectKey, exp)
	base := s.publicBaseURL
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/api/v1/files/local/%s?key=%s&exp=%d&sig=%s",
		base, op, url.QueryEscape(objectKey), exp, sig)
}

func (s *LocalStore) PresignPut(ctx context.Context, objectKey, contentType string, ttl time.Duration) (string, error) {
	_ = ctx
	_ = contentType
	return s.signedURL("put", objectKey, ttl), nil
}

func (s *LocalStore) PresignGet(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	_ = ctx
	return s.signedURL("get", objectKey, ttl), nil
}

func (s *LocalStore) Put(ctx context.Context, objectKey, contentType string, body io.Reader, size int64) error {
	_ = ctx
	_ = contentType
	_ = size
	path, err := s.pathFor(objectKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, body)
	return err
}

func (s *LocalStore) Get(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	_ = ctx
	return s.Open(objectKey)
}

func (s *LocalStore) Delete(ctx context.Context, objectKey string) error {
	_ = ctx
	path, err := s.pathFor(objectKey)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *LocalStore) Open(objectKey string) (*os.File, error) {
	path, err := s.pathFor(objectKey)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func ParseExp(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
