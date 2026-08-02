package application_test

import (
	"context"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/orders/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/orders/domain"
	orderspersist "github.com/Dovud1997/Dovud/backend/internal/modules/orders/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderLifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&orderspersist.OrderModel{},
		&orderspersist.OrderLineModel{},
		&orderspersist.OrderStatusHistoryModel{},
	); err != nil {
		t.Fatal(err)
	}

	svc := application.NewService(orderspersist.NewOrderRepo(db))
	tenantID := uuid.New()
	customerID := uuid.New()
	productID := uuid.New()
	ctx := context.Background()

	clientReq := "req-1"
	order, err := svc.CreateOrder(ctx, tenantID, application.CreateOrderInput{
		CustomerID: customerID, Currency: "UZS", Status: domain.StatusDraft, ClientRequestID: &clientReq,
		Lines: []application.OrderLineInput{{ProductID: productID, Qty: 2, UnitPrice: 1000}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if order.GrandTotal != 2000 {
		t.Fatalf("grand total=%v", order.GrandTotal)
	}

	again, err := svc.CreateOrder(ctx, tenantID, application.CreateOrderInput{
		CustomerID: customerID, Currency: "UZS", Status: domain.StatusDraft, ClientRequestID: &clientReq,
		Lines: []application.OrderLineInput{{ProductID: productID, Qty: 2, UnitPrice: 1000}},
	})
	if err != nil || again.ID != order.ID {
		t.Fatalf("idempotent create failed: err=%v id=%v", err, again)
	}

	submitted, err := svc.Submit(ctx, tenantID, order.ID, nil)
	if err != nil || submitted.Status != domain.StatusSubmitted {
		t.Fatalf("submit: err=%v status=%v", err, submitted)
	}
	confirmed, err := svc.Confirm(ctx, tenantID, order.ID, nil)
	if err != nil || confirmed.Status != domain.StatusConfirmed {
		t.Fatalf("confirm: err=%v status=%v", err, confirmed)
	}
}

func TestOrderCreateRecordsSyncChange(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&orderspersist.OrderModel{},
		&orderspersist.OrderLineModel{},
		&orderspersist.OrderStatusHistoryModel{},
	); err != nil {
		t.Fatal(err)
	}
	rec := &recordingFanout{}
	svc := application.NewService(orderspersist.NewOrderRepo(db)).WithSync(rec)
	tenantID := uuid.New()
	ctx := context.Background()
	order, err := svc.CreateOrder(ctx, tenantID, application.CreateOrderInput{
		CustomerID: uuid.New(), Currency: "UZS",
		Lines: []application.OrderLineInput{{ProductID: uuid.New(), Qty: 1, UnitPrice: 10}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.calls) != 1 || rec.calls[0].entityType != "order" || rec.calls[0].entityID != order.ID.String() {
		t.Fatalf("fanout=%+v", rec.calls)
	}
}

type recordingFanout struct {
	calls []struct {
		entityType string
		entityID   string
	}
}

func (r *recordingFanout) RecordChange(_ context.Context, _ uuid.UUID, entityType, entityID string, _ int64, _ bool, _ any) error {
	r.calls = append(r.calls, struct {
		entityType string
		entityID   string
	}{entityType, entityID})
	return nil
}
