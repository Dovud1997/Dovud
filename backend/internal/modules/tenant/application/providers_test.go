package application_test

import (
	"context"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/tenant/application"
	tenantpersist "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/platform/crypto"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUpsertAndMaskProvider(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&tenantpersist.ProviderModel{}); err != nil {
		t.Fatal(err)
	}
	box, err := crypto.NewSecretBox("test-access-secret-32-characters!!")
	if err != nil {
		t.Fatal(err)
	}
	svc := application.NewTenantService(
		tenantpersist.NewTenantRepo(db),
		tenantpersist.NewBrandingRepo(db),
		tenantpersist.NewDomainRepo(db),
	).WithProviders(tenantpersist.NewProviderRepo(db), box)

	tenantID := uuid.New()
	ctx := context.Background()
	dto, err := svc.UpsertProvider(ctx, tenantID, "smtp", application.UpsertProviderInput{
		Driver: "file",
		Config: map[string]any{
			"from": "demo@sfa.local", "password": "super-secret", "file_dir": t.TempDir(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if dto.Config["password"] != "********" {
		t.Fatalf("password not masked: %#v", dto.Config)
	}
	if err := svc.TestProvider(ctx, tenantID, "smtp", application.TestProviderInput{To: "a@b.c"}); err != nil {
		t.Fatal(err)
	}
}
