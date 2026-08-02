package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/analytics/application"
	analyticspersist "github.com/Dovud1997/Dovud/backend/internal/modules/analytics/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestKPISeedAndDashboard(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&analyticspersist.KpiDefinitionModel{},
		&analyticspersist.KpiSnapshotModel{},
	); err != nil {
		t.Fatal(err)
	}

	// Minimal stub tables for dashboard queries
	type orderRow struct {
		ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
		TenantID   uuid.UUID
		GrandTotal float64
		Status     string
		OrderedAt  time.Time
		DeletedAt  gorm.DeletedAt
	}
	type visitRow struct {
		ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
		TenantID  uuid.UUID
		StartedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	type agentRow struct {
		ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
		TenantID  uuid.UUID
		Status    string
		DeletedAt gorm.DeletedAt
	}
	if err := db.Table("orders").AutoMigrate(&orderRow{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Table("visits").AutoMigrate(&visitRow{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Table("sales_agents").AutoMigrate(&agentRow{}); err != nil {
		t.Fatal(err)
	}

	svc := application.NewService(analyticspersist.NewKpiRepo(db), db)
	tenantID := uuid.New()
	ctx := context.Background()

	defs, err := svc.ListKPIDefinitions(ctx, tenantID)
	if err != nil {
		t.Fatalf("list kpi: %v", err)
	}
	if len(defs) != 3 {
		t.Fatalf("expected 3 default KPIs, got %d", len(defs))
	}

	now := time.Now().UTC()
	_ = db.Table("orders").Create(&orderRow{
		ID: uuid.New(), TenantID: tenantID, GrandTotal: 1500, Status: "submitted", OrderedAt: now,
	})
	_ = db.Table("visits").Create(&visitRow{
		ID: uuid.New(), TenantID: tenantID, StartedAt: now,
	})
	_ = db.Table("sales_agents").Create(&agentRow{
		ID: uuid.New(), TenantID: tenantID, Status: "active",
	})

	summary, err := svc.DashboardSummary(ctx, tenantID)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if summary.OrdersToday != 1 || summary.OrdersTotalAmountToday != 1500 {
		t.Fatalf("orders today=%d amount=%v", summary.OrdersToday, summary.OrdersTotalAmountToday)
	}
	if summary.VisitsToday != 1 || summary.ActiveAgents != 1 || summary.PendingOrders != 1 {
		t.Fatalf("visits=%d agents=%d pending=%d", summary.VisitsToday, summary.ActiveAgents, summary.PendingOrders)
	}
	if summary.OpenReceivables != 0 {
		t.Fatalf("expected 0 receivables without table, got %v", summary.OpenReceivables)
	}

	from := now.Add(-24 * time.Hour)
	to := now.Add(24 * time.Hour)
	sales, err := svc.SalesAnalytics(ctx, tenantID, from, to)
	if err != nil || len(sales) == 0 {
		t.Fatalf("sales analytics: err=%v len=%d", err, len(sales))
	}
}
