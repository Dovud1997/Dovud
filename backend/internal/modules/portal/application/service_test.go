package application_test

import (
	"context"
	"testing"

	crmpersist "github.com/Dovud1997/Dovud/backend/internal/modules/crm/infrastructure/persistence"
	docspersist "github.com/Dovud1997/Dovud/backend/internal/modules/documents/infrastructure/persistence"
	documentsdomain "github.com/Dovud1997/Dovud/backend/internal/modules/documents/domain"
	financepersist "github.com/Dovud1997/Dovud/backend/internal/modules/finance/infrastructure/persistence"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	orderspersist "github.com/Dovud1997/Dovud/backend/internal/modules/orders/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/modules/portal/application"
	portaldomain "github.com/Dovud1997/Dovud/backend/internal/modules/portal/domain"
	portalpersist "github.com/Dovud1997/Dovud/backend/internal/modules/portal/infrastructure/persistence"
	crmdomain "github.com/Dovud1997/Dovud/backend/internal/modules/crm/domain"
	identitydomain "github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupPortal(t *testing.T) (*application.Service, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&identitypersist.UserModel{},
		&crmpersist.CustomerModel{},
		&orderspersist.OrderModel{},
		&orderspersist.OrderLineModel{},
		&orderspersist.OrderStatusHistoryModel{},
		&financepersist.ReceivableModel{},
		&financepersist.ReceivablePaymentModel{},
		&docspersist.DocumentModel{},
		&portalpersist.CustomerUserModel{},
	); err != nil {
		t.Fatal(err)
	}
	tenantID := uuid.New()
	userID := uuid.New()
	customerID := uuid.New()
	otherCustomer := uuid.New()

	userRepo := identitypersist.NewUserRepo(db)
	_ = userRepo.Create(context.Background(), &identitydomain.User{
		ID: userID, TenantID: tenantID, Email: "p@test.local", PasswordHash: "x",
		FullName: "Portal", Locale: "en", ThemePreference: "system", Status: "active",
	})
	custRepo := crmpersist.NewCustomerRepo(db)
	_ = custRepo.Create(context.Background(), &crmdomain.Customer{
		ID: customerID, TenantID: tenantID, Code: "C1", Name: "Cust One", Type: "outlet", Status: "active", Version: 1,
	})
	_ = custRepo.Create(context.Background(), &crmdomain.Customer{
		ID: otherCustomer, TenantID: tenantID, Code: "C2", Name: "Cust Two", Type: "outlet", Status: "active", Version: 1,
	})
	docRepo := docspersist.NewDocumentRepo(db)
	_ = docRepo.Create(context.Background(), &documentsdomain.Document{
		TenantID: tenantID, CustomerID: &customerID, Title: "Mine", DocType: "general", Status: "active",
	})
	_ = docRepo.Create(context.Background(), &documentsdomain.Document{
		TenantID: tenantID, CustomerID: &otherCustomer, Title: "Other", DocType: "general", Status: "active",
	})
	linkRepo := portalpersist.NewCustomerUserRepo(db)
	_ = linkRepo.Upsert(context.Background(), &portaldomain.CustomerUser{
		TenantID: tenantID, UserID: userID, CustomerID: customerID,
	})

	svc := application.NewService(
		linkRepo, custRepo,
		orderspersist.NewOrderRepo(db),
		financepersist.NewReceivableRepo(db),
		docRepo, userRepo,
	)
	return svc, tenantID, userID, customerID
}

func TestPortalDocumentsScopedToCustomer(t *testing.T) {
	svc, tenantID, userID, _ := setupPortal(t)
	docs, total, err := svc.Documents(context.Background(), tenantID, userID, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(docs) != 1 || docs[0].Title != "Mine" {
		t.Fatalf("got total=%d docs=%v", total, docs)
	}
	sum, err := svc.Summary(context.Background(), tenantID, userID)
	if err != nil {
		t.Fatal(err)
	}
	if sum.DocumentsCount != 1 {
		t.Fatalf("documents_count=%d", sum.DocumentsCount)
	}
}
