package application_test

import (
	"context"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/crm/application"
	crmpersist "github.com/Dovud1997/Dovud/backend/internal/modules/crm/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCustomerWithContact(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&crmpersist.CustomerCategoryModel{},
		&crmpersist.CustomerModel{},
		&crmpersist.CustomerContactModel{},
		&crmpersist.CustomerAddressModel{},
	); err != nil {
		t.Fatal(err)
	}
	svc := application.NewService(
		crmpersist.NewCustomerRepo(db),
		crmpersist.NewCustomerContactRepo(db),
		crmpersist.NewCustomerAddressRepo(db),
		crmpersist.NewCustomerCategoryRepo(db),
	)
	tenantID := uuid.New()
	ctx := context.Background()

	limit := 1000.0
	cust, err := svc.CreateCustomer(ctx, tenantID, application.CustomerInput{
		Code: "C1", Name: "Market", Type: "outlet", CreditLimit: &limit,
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}
	contact, err := svc.CreateContact(ctx, tenantID, cust.ID, application.ContactInput{
		FullName: "Ali", Phone: "+99890", IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("create contact: %v", err)
	}
	if contact.FullName != "Ali" {
		t.Fatalf("unexpected contact %+v", contact)
	}
	list, total, err := svc.ListCustomers(ctx, tenantID, 1, 20)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list customers: err=%v total=%d len=%d", err, total, len(list))
	}
}
