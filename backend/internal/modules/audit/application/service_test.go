package application_test

import (
	"context"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/audit/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/audit/domain"
	auditpersist "github.com/Dovud1997/Dovud/backend/internal/modules/audit/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWriteAndList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	if err := db.AutoMigrate(&auditpersist.AuditLogModel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	svc := application.NewService(auditpersist.NewAuditRepo(db), nil)
	tenantID := uuid.New()
	actor := uuid.New()
	entityID := uuid.New()
	ctx := context.Background()

	row, err := svc.Write(ctx, application.WriteInput{
		TenantID: &tenantID, ActorUserID: &actor, Action: "order.status_changed",
		EntityType: "order", EntityID: &entityID,
		Before: map[string]any{"status": "submitted"},
		After:  map[string]any{"status": "confirmed"},
		RequestID: "req-1",
	})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if row.Action != "order.status_changed" {
		t.Fatalf("action=%s", row.Action)
	}

	list, total, err := svc.List(ctx, tenantID, domain.ListFilters{EntityType: "order", EntityID: &entityID}, 1, 20)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("total=%d len=%d", total, len(list))
	}
	got, err := svc.Get(ctx, tenantID, row.ID)
	if err != nil || got.ID != row.ID {
		t.Fatalf("get: %v", err)
	}
}
