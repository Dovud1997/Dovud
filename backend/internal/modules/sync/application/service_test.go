package application_test

import (
	"context"
	"testing"
	"time"

	crmdomain "github.com/Dovud1997/Dovud/backend/internal/modules/crm/domain"
	crmpersist "github.com/Dovud1997/Dovud/backend/internal/modules/crm/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/domain"
	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/infrastructure/applicator"
	syncpersist "github.com/Dovud1997/Dovud/backend/internal/modules/sync/infrastructure/persistence"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
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

func TestConflictClientWinsAppearsOnPull(t *testing.T) {
	svc := setupSyncService(t)
	tenantID := uuid.New()
	userID := uuid.New()
	deviceID := "device-conflict"
	entityID := uuid.New().String()
	ctx := context.Background()

	if _, err := svc.Bootstrap(ctx, tenantID, userID, application.BootstrapInput{DeviceID: deviceID}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Push(ctx, tenantID, userID, application.PushInput{
		DeviceID: deviceID,
		Ops: []domain.SyncOp{{
			OpID: uuid.New().String(), EntityType: "order", EntityID: entityID,
			Op: domain.OpCreate, Payload: map[string]any{"number": "A"}, ClientTS: time.Now().UTC(),
		}},
	}); err != nil {
		t.Fatal(err)
	}
	// Simulate concurrent server edit by appending another version.
	if err := svc.RecordChange(ctx, tenantID, "order", entityID, 0, false, map[string]any{"number": "SERVER"}); err != nil {
		t.Fatal(err)
	}
	push, err := svc.Push(ctx, tenantID, userID, application.PushInput{
		DeviceID: deviceID,
		Ops: []domain.SyncOp{{
			OpID: uuid.New().String(), EntityType: "order", EntityID: entityID,
			Op: domain.OpUpdate, BaseVersion: 1, Payload: map[string]any{"number": "MINE"}, ClientTS: time.Now().UTC(),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(push.Results) != 1 || push.Results[0].Status != domain.PushConflict || push.Results[0].ConflictID == nil {
		t.Fatalf("expected conflict, got %+v", push.Results)
	}
	conflictID := *push.Results[0].ConflictID
	open, err := svc.ListConflicts(ctx, tenantID, deviceID)
	if err != nil || len(open) != 1 {
		t.Fatalf("open=%v err=%v", open, err)
	}
	resolved, err := svc.ResolveConflict(ctx, tenantID, conflictID, application.ResolveConflictInput{
		Resolution: domain.ResolutionClientWins,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Status != domain.ConflictStatusResolved || resolved.Resolution == nil || *resolved.Resolution != domain.ResolutionClientWins {
		t.Fatalf("resolved=%+v", resolved)
	}
	pull, err := svc.Pull(ctx, tenantID, userID, deviceID, "", []string{"order"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	var last map[string]any
	for _, ch := range pull.Changes {
		if ch.EntityID == entityID {
			last, _ = ch.Data.(map[string]any)
		}
	}
	if last == nil || last["number"] != "MINE" {
		t.Fatalf("expected client payload on pull, got %v (changes=%d)", last, len(pull.Changes))
	}
}

func TestPushRespectsDeviceLock(t *testing.T) {
	svc := setupSyncService(t)
	locker := &memLocker{held: map[string]string{}}
	svc.WithLocker(locker)
	tenantID := uuid.New()
	userID := uuid.New()
	deviceID := "locked-device"
	ctx := context.Background()
	if _, err := svc.Bootstrap(ctx, tenantID, userID, application.BootstrapInput{DeviceID: deviceID}); err != nil {
		t.Fatal(err)
	}
	key := "sync:lock:" + tenantID.String() + ":" + deviceID
	locker.held[key] = "other"
	_, err := svc.Push(ctx, tenantID, userID, application.PushInput{
		DeviceID: deviceID,
		Ops: []domain.SyncOp{{
			OpID: uuid.New().String(), EntityType: "note", EntityID: "n1",
			Op: domain.OpCreate, Payload: map[string]any{"t": 1}, ClientTS: time.Now().UTC(),
		}},
	})
	if err != apperrors.ErrRateLimited {
		t.Fatalf("expected rate limited, got %v", err)
	}
}

type memLocker struct {
	held map[string]string
}

func (m *memLocker) TryLock(_ context.Context, key, token string, _ time.Duration) (bool, error) {
	if _, ok := m.held[key]; ok {
		return false, nil
	}
	m.held[key] = token
	return true, nil
}

func (m *memLocker) Unlock(_ context.Context, key, token string) error {
	if m.held[key] == token {
		delete(m.held, key)
	}
	return nil
}

func TestPushCreateCustomerAppliesDomain(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&syncpersist.SyncDeviceModel{},
		&syncpersist.SyncChangeLogModel{},
		&syncpersist.SyncConflictModel{},
		&syncpersist.SyncAppliedOpModel{},
		&crmpersist.CustomerModel{},
	); err != nil {
		t.Fatal(err)
	}
	svc := application.NewService(
		syncpersist.NewDeviceRepo(db),
		syncpersist.NewChangeLogRepo(db),
		syncpersist.NewConflictRepo(db),
	).WithApplicator(applicator.New(
		crmpersist.NewCustomerRepo(db),
		nil,
		nil,
	))
	// orders/visits nil — only customer supported path used
	tenantID := uuid.New()
	userID := uuid.New()
	deviceID := "dev-apply"
	entityID := uuid.New().String()
	ctx := context.Background()
	if _, err := svc.Bootstrap(ctx, tenantID, userID, application.BootstrapInput{DeviceID: deviceID}); err != nil {
		t.Fatal(err)
	}
	push, err := svc.Push(ctx, tenantID, userID, application.PushInput{
		DeviceID: deviceID,
		Ops: []domain.SyncOp{{
			OpID: uuid.New().String(), EntityType: "customer", EntityID: entityID,
			Op: domain.OpCreate, Payload: map[string]any{"code": "C1", "name": "Acme"}, ClientTS: time.Now().UTC(),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(push.Results) != 1 || push.Results[0].Status != domain.PushAcked {
		t.Fatalf("push=%+v", push.Results)
	}
	cid, _ := uuid.Parse(entityID)
	var row crmpersist.CustomerModel
	if err := db.Where("id = ? AND tenant_id = ?", cid, tenantID).First(&row).Error; err != nil {
		t.Fatalf("customer missing: %v", err)
	}
	if row.Name != "Acme" || row.Code != "C1" {
		t.Fatalf("customer=%+v", row)
	}
}

func TestResolveConflictMergeAppliesDomain(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&syncpersist.SyncDeviceModel{},
		&syncpersist.SyncChangeLogModel{},
		&syncpersist.SyncConflictModel{},
		&syncpersist.SyncAppliedOpModel{},
		&crmpersist.CustomerModel{},
	); err != nil {
		t.Fatal(err)
	}
	customers := crmpersist.NewCustomerRepo(db)
	cl := syncpersist.NewChangeLogRepo(db)
	svc := application.NewService(
		syncpersist.NewDeviceRepo(db),
		cl,
		syncpersist.NewConflictRepo(db),
	).WithApplicator(applicator.New(customers, nil, nil))

	tenantID := uuid.New()
	userID := uuid.New()
	deviceID := "dev-merge"
	entityID := uuid.New()
	ctx := context.Background()
	if _, err := svc.Bootstrap(ctx, tenantID, userID, application.BootstrapInput{DeviceID: deviceID}); err != nil {
		t.Fatal(err)
	}

	addr := "Main St"
	c := &crmdomain.Customer{
		ID: entityID, TenantID: tenantID, Code: "MX", Name: "ServerV2", Type: "outlet", Status: "active",
		Address: &addr, Version: 1,
	}
	if err := customers.Create(ctx, c); err != nil {
		t.Fatal(err)
	}
	// Simulate server already at changelog v2
	if err := cl.Append(ctx, &domain.SyncChange{
		TenantID: tenantID, EntityType: "customer", EntityID: entityID.String(),
		Version: 1, PayloadJSON: `{"name":"ServerName","code":"MX"}`,
	}); err != nil {
		t.Fatal(err)
	}
	if err := customers.Update(ctx, c); err != nil {
		t.Fatal(err)
	}
	if err := cl.Append(ctx, &domain.SyncChange{
		TenantID: tenantID, EntityType: "customer", EntityID: entityID.String(),
		Version: 2, PayloadJSON: `{"name":"ServerV2","code":"MX","address":"Main St"}`,
	}); err != nil {
		t.Fatal(err)
	}

	push, err := svc.Push(ctx, tenantID, userID, application.PushInput{
		DeviceID: deviceID,
		Ops: []domain.SyncOp{{
			OpID: uuid.New().String(), EntityType: "customer", EntityID: entityID.String(),
			Op: domain.OpUpdate, BaseVersion: 1,
			Payload:  map[string]any{"name": "ClientName", "code": "MX", "address": "Client Ave"},
			ClientTS: time.Now().UTC(),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(push.Results) != 1 || push.Results[0].Status != domain.PushConflict || push.Results[0].ConflictID == nil {
		t.Fatalf("expected conflict, got %+v", push.Results)
	}

	resolved, err := svc.ResolveConflict(ctx, tenantID, *push.Results[0].ConflictID, application.ResolveConflictInput{
		Resolution: domain.ResolutionMerge,
		MergedPayload: map[string]any{
			"name": "ClientName", "code": "MX", "address": "Main St",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Status != domain.ConflictStatusResolved || resolved.Resolution == nil || *resolved.Resolution != domain.ResolutionMerge {
		t.Fatalf("resolved=%+v", resolved)
	}
	var row crmpersist.CustomerModel
	if err := db.Where("id = ?", entityID).First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.Name != "ClientName" {
		t.Fatalf("expected merged name ClientName, got %+v", row)
	}
	if row.Address == nil || *row.Address != "Main St" {
		t.Fatalf("expected server address kept, got %+v", row)
	}
}

func strPtr(s string) *string { return &s }
