package application_test

import (
	"context"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/catalog/application"
	catalogpersist "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type memRecorder struct {
	calls []struct {
		EntityType string
		EntityID   string
		Version    int64
		Deleted    bool
	}
}

func (m *memRecorder) RecordChange(_ context.Context, _ uuid.UUID, entityType, entityID string, version int64, deleted bool, _ any) error {
	m.calls = append(m.calls, struct {
		EntityType string
		EntityID   string
		Version    int64
		Deleted    bool
	}{entityType, entityID, version, deleted})
	return nil
}

func TestCreateProductAndPrice(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&catalogpersist.ManufacturerModel{},
		&catalogpersist.CategoryModel{},
		&catalogpersist.ProductModel{},
		&catalogpersist.PriceListModel{},
		&catalogpersist.ProductPriceModel{},
		&catalogpersist.PromotionModel{},
		&catalogpersist.PromotionItemModel{},
	); err != nil {
		t.Fatal(err)
	}

	rec := &memRecorder{}
	svc := application.NewService(
		catalogpersist.NewManufacturerRepo(db),
		catalogpersist.NewCategoryRepo(db),
		catalogpersist.NewProductRepo(db),
		catalogpersist.NewPriceRepo(db),
		catalogpersist.NewPromotionRepo(db),
	).WithSync(rec)
	tenantID := uuid.New()
	ctx := context.Background()

	m, err := svc.CreateManufacturer(ctx, tenantID, application.ManufacturerInput{Code: "ACME", Name: "Acme"})
	if err != nil {
		t.Fatalf("manufacturer: %v", err)
	}
	c, err := svc.CreateCategory(ctx, tenantID, application.CategoryInput{Code: "BEV", Name: "Beverages"})
	if err != nil {
		t.Fatalf("category: %v", err)
	}
	p, err := svc.CreateProduct(ctx, tenantID, application.ProductInput{
		SKU: "SKU-1", Name: "Cola", CategoryID: &c.ID, ManufacturerID: &m.ID, Unit: "pcs", VATRate: 12,
	})
	if err != nil {
		t.Fatalf("product: %v", err)
	}
	pl, err := svc.CreatePriceList(ctx, tenantID, application.PriceListInput{Code: "STD", Name: "Standard", Currency: "UZS", IsDefault: true})
	if err != nil {
		t.Fatalf("price list: %v", err)
	}
	price, err := svc.UpsertPrice(ctx, tenantID, pl.ID, application.PriceInput{ProductID: p.ID, Amount: 1000, Currency: "UZS"})
	if err != nil {
		t.Fatalf("price: %v", err)
	}
	if price.Amount != 1000 {
		t.Fatalf("unexpected amount %v", price.Amount)
	}
	if len(rec.calls) < 2 {
		t.Fatalf("expected product+price sync fan-out, got %+v", rec.calls)
	}
	if rec.calls[0].EntityType != "product" || rec.calls[1].EntityType != "product_price" {
		t.Fatalf("fan-out order %+v", rec.calls)
	}
}
