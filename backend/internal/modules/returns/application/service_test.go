package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/returns/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/returns/domain"
	returnspersist "github.com/Dovud1997/Dovud/backend/internal/modules/returns/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestReturnCreateSubmitApprove(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&returnspersist.ReturnModel{},
		&returnspersist.ReturnLineModel{},
	); err != nil {
		t.Fatal(err)
	}

	svc := application.NewService(returnspersist.NewReturnRepo(db))
	tenantID := uuid.New()
	customerID := uuid.New()
	productID := uuid.New()
	ctx := context.Background()

	ret, err := svc.CreateReturn(ctx, tenantID, application.CreateReturnInput{
		CustomerID: customerID, Currency: "UZS", Status: domain.StatusDraft,
		Lines: []application.ReturnLineInput{{ProductID: productID, Qty: 2, UnitPrice: 500}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if ret.GrandTotal != 1000 {
		t.Fatalf("grand total=%v", ret.GrandTotal)
	}
	if !strings.HasPrefix(ret.Number, "RET-") {
		t.Fatalf("number=%v", ret.Number)
	}

	submitted, err := svc.Submit(ctx, tenantID, ret.ID)
	if err != nil || submitted.Status != domain.StatusSubmitted {
		t.Fatalf("submit: err=%v status=%v", err, submitted)
	}

	approved, err := svc.Approve(ctx, tenantID, ret.ID)
	if err != nil || approved.Status != domain.StatusApproved {
		t.Fatalf("approve: err=%v status=%v", err, approved)
	}
}
