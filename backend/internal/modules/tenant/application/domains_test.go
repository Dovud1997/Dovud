package application_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	docsapp "github.com/Dovud1997/Dovud/backend/internal/modules/documents/application"
	docspersist "github.com/Dovud1997/Dovud/backend/internal/modules/documents/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/modules/tenant/application"
	tenantpersist "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/Dovud1997/Dovud/backend/internal/platform/storage"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTenantDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:domains-"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&tenantpersist.TenantModel{},
		&tenantpersist.BrandingModel{},
		&tenantpersist.DomainModel{},
		&docspersist.FileModel{},
		&docspersist.DocumentModel{},
		&docspersist.DocumentFileModel{},
		&outbox.EventModel{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestAddDomainNormalizePrimaryAndConflict(t *testing.T) {
	db := setupTenantDB(t)
	svc := application.NewTenantService(
		tenantpersist.NewTenantRepo(db),
		tenantpersist.NewBrandingRepo(db),
		tenantpersist.NewDomainRepo(db),
	)
	tenantID := uuid.New()
	ctx := context.Background()

	first, err := svc.AddDomain(ctx, tenantID, application.CreateDomainInput{
		Host: "https://Demo.Example.com/path", IsPrimary: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Host != "demo.example.com" || !first.IsPrimary {
		t.Fatalf("unexpected first domain: %+v", first)
	}

	second, err := svc.AddDomain(ctx, tenantID, application.CreateDomainInput{
		Host: "app.example.com", IsPrimary: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := svc.ListDomains(ctx, tenantID)
	if err != nil {
		t.Fatal(err)
	}
	var primaryCount int
	for _, d := range list {
		if d.IsPrimary {
			primaryCount++
			if d.ID != second.ID {
				t.Fatalf("expected second domain to be primary, got %+v", d)
			}
		}
	}
	if primaryCount != 1 {
		t.Fatalf("expected one primary, got %d", primaryCount)
	}

	if _, err := svc.AddDomain(ctx, tenantID, application.CreateDomainInput{Host: "demo.example.com"}); err == nil {
		t.Fatal("expected domain conflict")
	}
	if _, err := svc.AddDomain(ctx, tenantID, application.CreateDomainInput{Host: "not a host!!"}); err == nil {
		t.Fatal("expected invalid host")
	}
}

func TestAttachBrandingAssetFromReadyImage(t *testing.T) {
	db := setupTenantDB(t)
	dir := t.TempDir()
	store, err := storage.NewLocalStore(filepath.Join(dir, "storage"), "sfa", "http://localhost:8080", "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	docsSvc := docsapp.NewService(
		docspersist.NewFileRepo(db),
		docspersist.NewDocumentRepo(db),
		store,
		outbox.NewStore(db),
	)
	svc := application.NewTenantService(
		tenantpersist.NewTenantRepo(db),
		tenantpersist.NewBrandingRepo(db),
		tenantpersist.NewDomainRepo(db),
	).WithAssets(docsSvc)

	tenantID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	presign, err := docsSvc.PresignUpload(ctx, tenantID, userID, docsapp.PresignInput{
		FileName: "logo.png", Mime: "image/png", Size: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	body := []byte{0x89, 0x50, 0x4e, 0x47}
	if err := store.Put(ctx, presign.ObjectKey, "image/png", bytes.NewReader(body), int64(len(body))); err != nil {
		t.Fatal(err)
	}
	size := int64(len(body))
	if _, err := docsSvc.CompleteUpload(ctx, tenantID, presign.FileID, docsapp.CompleteInput{Size: &size}); err != nil {
		t.Fatal(err)
	}

	branding, err := svc.AttachBrandingAsset(ctx, tenantID, application.AttachBrandingAssetInput{
		FileID: presign.FileID, Kind: "logo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if branding.LogoURL == nil || *branding.LogoURL == "" {
		t.Fatal("expected logo_url")
	}

	// non-image should fail
	presign2, err := docsSvc.PresignUpload(ctx, tenantID, userID, docsapp.PresignInput{
		FileName: "notes.txt", Mime: "text/plain", Size: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = store.Put(ctx, presign2.ObjectKey, "text/plain", bytes.NewReader([]byte("x")), 1)
	sz := int64(1)
	_, _ = docsSvc.CompleteUpload(ctx, tenantID, presign2.FileID, docsapp.CompleteInput{Size: &sz})
	if _, err := svc.AttachBrandingAsset(ctx, tenantID, application.AttachBrandingAssetInput{
		FileID: presign2.FileID, Kind: "logo",
	}); err == nil {
		t.Fatal("expected non-image rejection")
	}
}
