package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/domain"
	syncpersist "github.com/Dovud1997/Dovud/backend/internal/modules/sync/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSyncService(t *testing.T) *application.Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&syncpersist.SyncDeviceModel{},
		&syncpersist.SyncChangeLogModel{},
		&syncpersist.SyncConflictModel{},
		&syncpersist.SyncAppliedOpModel{},
	); err != nil {
		t.Fatal(err)
	}
	return application.NewService(
		syncpersist.NewDeviceRepo(db),
		syncpersist.NewChangeLogRepo(db),
		syncpersist.NewConflictRepo(db),
	)
}

func TestBootstrapPushCreatePullSeesChange(t *testing.T) {
	svc := setupSyncService(t)
	tenantID := uuid.New()
	userID := uuid.New()
	deviceID := "device-1"
	entityID := uuid.New().String()
	ctx := context.Background()

	boot, err := svc.Bootstrap(ctx, tenantID, userID, application.BootstrapInput{
		DeviceID: deviceID, Platform: strPtr("android"), AppVersion: strPtr("1.0.0"),
	})
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if boot.SyncProtocol != domain.SyncProtocol {
		t.Fatalf("protocol=%d", boot.SyncProtocol)
	}

	push, err := svc.Push(ctx, tenantID, userID, application.PushInput{
		DeviceID: deviceID,
		Ops: []domain.SyncOp{{
			OpID: uuid.New().String(), EntityType: "order", EntityID: entityID,
			Op: domain.OpCreate, BaseVersion: 0,
			Payload:  map[string]any{"number": "ORD-1"},
			ClientTS: time.Now().UTC(),
		}},
	})
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	if len(push.Results) != 1 || push.Results[0].Status != domain.PushAcked {
		t.Fatalf("push results=%+v", push.Results)
	}
	if push.Results[0].Version == nil || *push.Results[0].Version != 1 {
		t.Fatalf("version=%v", push.Results[0].Version)
	}

	pull, err := svc.Pull(ctx, tenantID, userID, deviceID, "", []string{"order"}, 50)
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if len(pull.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(pull.Changes))
	}
	ch := pull.Changes[0]
	if ch.EntityType != "order" || ch.EntityID != entityID || ch.Version != 1 || ch.Deleted {
		t.Fatalf("change=%+v", ch)
	}
	data, ok := ch.Data.(map[string]any)
	if !ok || data["number"] != "ORD-1" {
		t.Fatalf("data=%v", ch.Data)
	}
	if pull.NextCursor == "" {
		t.Fatal("expected next_cursor")
	}
}

func strPtr(s string) *string { return &s }
