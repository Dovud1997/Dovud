package application_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/documents/application"
	docspersist "github.com/Dovud1997/Dovud/backend/internal/modules/documents/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/Dovud1997/Dovud/backend/internal/platform/storage"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setup(t *testing.T) (*application.Service, *storage.LocalStore) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	if err := db.AutoMigrate(
		&docspersist.FileModel{},
		&docspersist.DocumentModel{},
		&docspersist.DocumentFileModel{},
		&outbox.EventModel{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	dir := t.TempDir()
	store, err := storage.NewLocalStore(filepath.Join(dir, "storage"), "sfa", "http://localhost:8080", "test-secret")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	svc := application.NewService(
		docspersist.NewFileRepo(db),
		docspersist.NewDocumentRepo(db),
		store,
		outbox.NewStore(db),
	)
	return svc, store
}

func TestPresignUploadCompleteAndAttach(t *testing.T) {
	svc, store := setup(t)
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()

	presign, err := svc.PresignUpload(ctx, tenantID, userID, application.PresignInput{
		FileName: "note.txt", Mime: "text/plain", Size: 5,
	})
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	if presign.UploadURL == "" || presign.FileID == uuid.Nil {
		t.Fatal("expected upload url and file id")
	}
	if store.Driver() != "local" {
		t.Fatalf("driver=%s", store.Driver())
	}

	body := []byte("hello")
	if err := store.Put(ctx, presign.ObjectKey, "text/plain", bytes.NewReader(body), int64(len(body))); err != nil {
		t.Fatalf("put: %v", err)
	}

	size := int64(len(body))
	file, err := svc.CompleteUpload(ctx, tenantID, presign.FileID, application.CompleteInput{Size: &size})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if file.Status != "ready" {
		t.Fatalf("status=%s", file.Status)
	}

	doc, err := svc.CreateDocument(ctx, tenantID, userID, application.CreateDocumentInput{
		Title: "Demo contract", DocType: "contract",
	})
	if err != nil {
		t.Fatalf("create doc: %v", err)
	}
	attached, err := svc.AttachFile(ctx, tenantID, doc.ID, application.AttachFileInput{FileID: file.ID})
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	if len(attached.Files) != 1 {
		t.Fatalf("files=%d", len(attached.Files))
	}

	f, err := store.Open(presign.ObjectKey)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = f.Close()
}
