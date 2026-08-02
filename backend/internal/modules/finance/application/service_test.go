package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/finance/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/finance/domain"
	financepersist "github.com/Dovud1997/Dovud/backend/internal/modules/finance/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupService(t *testing.T) *application.Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&financepersist.ReceivableModel{},
		&financepersist.ReceivablePaymentModel{},
		&financepersist.CreditLimitModel{},
	); err != nil {
		t.Fatal(err)
	}
	return application.NewService(
		financepersist.NewReceivableRepo(db),
		financepersist.NewCreditLimitRepo(db),
	)
}

func TestCreateReceivableAndPaymentClosesIt(t *testing.T) {
	svc := setupService(t)
	tenantID := uuid.New()
	customerID := uuid.New()
	ctx := context.Background()

	due := time.Now().UTC().Add(7 * 24 * time.Hour)
	rec, err := svc.CreateReceivable(ctx, tenantID, application.CreateReceivableInput{
		CustomerID: customerID, Amount: 1000, Currency: "UZS", DueDate: &due,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if rec.DocumentType != domain.DocumentTypeManual {
		t.Fatalf("document_type=%s", rec.DocumentType)
	}
	if rec.Balance != 1000 || rec.PaidAmount != 0 || rec.Status != domain.StatusOpen {
		t.Fatalf("unexpected receivable %+v", rec)
	}

	partial, err := svc.RecordPayment(ctx, tenantID, rec.ID, application.RecordPaymentInput{
		Amount: 400, Method: domain.MethodCash,
	}, nil)
	if err != nil {
		t.Fatalf("partial payment: %v", err)
	}
	if partial.Status != domain.StatusPartial || partial.PaidAmount != 400 || partial.Balance != 600 {
		t.Fatalf("partial state %+v", partial)
	}

	closed, err := svc.RecordPayment(ctx, tenantID, rec.ID, application.RecordPaymentInput{
		Amount: 600, Method: domain.MethodTransfer,
	}, nil)
	if err != nil {
		t.Fatalf("closing payment: %v", err)
	}
	if closed.Status != domain.StatusClosed || closed.Balance != 0 || closed.PaidAmount != 1000 {
		t.Fatalf("closed state %+v", closed)
	}
	if len(closed.Payments) != 2 {
		t.Fatalf("payments=%d", len(closed.Payments))
	}

	bal, err := svc.GetCustomerBalance(ctx, tenantID, customerID)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if bal.Balance != 0 {
		t.Fatalf("customer balance=%v", bal.Balance)
	}
}
