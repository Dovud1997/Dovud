package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/application"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	analyticspersist "github.com/Dovud1997/Dovud/backend/internal/modules/analytics/infrastructure/persistence"
	auditpersist "github.com/Dovud1997/Dovud/backend/internal/modules/audit/infrastructure/persistence"
	catalogpersist "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/infrastructure/persistence"
	crmpersist "github.com/Dovud1997/Dovud/backend/internal/modules/crm/infrastructure/persistence"
	docspersist "github.com/Dovud1997/Dovud/backend/internal/modules/documents/infrastructure/persistence"
	ffpersist "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/infrastructure/persistence"
	financepersist "github.com/Dovud1997/Dovud/backend/internal/modules/finance/infrastructure/persistence"
	notifypersist "github.com/Dovud1997/Dovud/backend/internal/modules/notifications/infrastructure/persistence"
	orderspersist "github.com/Dovud1997/Dovud/backend/internal/modules/orders/infrastructure/persistence"
	orgpersist "github.com/Dovud1997/Dovud/backend/internal/modules/organization/infrastructure/persistence"
	portalpersist "github.com/Dovud1997/Dovud/backend/internal/modules/portal/infrastructure/persistence"
	returnspersist "github.com/Dovud1997/Dovud/backend/internal/modules/returns/infrastructure/persistence"
	syncpersist "github.com/Dovud1997/Dovud/backend/internal/modules/sync/infrastructure/persistence"
	tenantpersist "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/seed"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	if err := db.AutoMigrate(
		&tenantpersist.TenantModel{},
		&tenantpersist.BrandingModel{},
		&tenantpersist.DomainModel{},
		&identitypersist.PermissionModel{},
		&identitypersist.RoleModel{},
		&identitypersist.RolePermissionModel{},
		&identitypersist.UserModel{},
		&identitypersist.UserRoleModel{},
		&identitypersist.RefreshTokenModel{},
		&identitypersist.UserDeviceModel{},
		&orgpersist.CompanyModel{},
		&orgpersist.BranchModel{},
		&orgpersist.WarehouseModel{},
		&orgpersist.WarehouseStockModel{},
		&catalogpersist.ManufacturerModel{},
		&catalogpersist.CategoryModel{},
		&catalogpersist.ProductModel{},
		&catalogpersist.PriceListModel{},
		&catalogpersist.ProductPriceModel{},
		&catalogpersist.PromotionModel{},
		&catalogpersist.PromotionItemModel{},
		&crmpersist.CustomerCategoryModel{},
		&crmpersist.CustomerModel{},
		&crmpersist.CustomerContactModel{},
		&crmpersist.CustomerAddressModel{},
		&ffpersist.SalesAgentModel{},
		&ffpersist.RouteModel{},
		&ffpersist.RouteStopModel{},
		&ffpersist.VisitModel{},
		&ffpersist.VisitPhotoModel{},
		&ffpersist.VisitCommentModel{},
		&ffpersist.GpsTrackModel{},
		&orderspersist.OrderModel{},
		&orderspersist.OrderLineModel{},
		&orderspersist.OrderStatusHistoryModel{},
		&returnspersist.ReturnModel{},
		&returnspersist.ReturnLineModel{},
		&financepersist.ReceivableModel{},
		&financepersist.ReceivablePaymentModel{},
		&financepersist.CreditLimitModel{},
		&syncpersist.SyncDeviceModel{},
		&syncpersist.SyncChangeLogModel{},
		&syncpersist.SyncConflictModel{},
		&syncpersist.SyncAppliedOpModel{},
		&notifypersist.NotificationModel{},
		&notifypersist.NotificationDeliveryModel{},
		&analyticspersist.KpiDefinitionModel{},
		&analyticspersist.KpiSnapshotModel{},
		&outbox.EventModel{},
		&docspersist.FileModel{},
		&docspersist.DocumentModel{},
		&docspersist.DocumentFileModel{},
		&docspersist.EntityFileModel{},
		&auditpersist.AuditLogModel{},
		&portalpersist.CustomerUserModel{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := seed.Run(context.Background(), db, nil); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

func TestLoginRefreshLogout(t *testing.T) {
	db := setupTestDB(t)
	tokenSvc := auth.NewTokenService(config.AuthConfig{
		AccessSecret:     "test-access-secret-32-characters!!",
		RefreshSecret:    "test-refresh-secret-32-characters!",
		AccessTTLMinutes: 15,
		RefreshTTLDays:   30,
		Issuer:           "sfa-test",
	})

	authSvc := application.NewAuthService(
		identitypersist.NewUserRepo(db),
		identitypersist.NewRefreshTokenRepo(db),
		identitypersist.NewDeviceRepo(db),
		tenantpersist.NewTenantRepo(db),
		tokenSvc,
	)

	ctx := context.Background()
	res, err := authSvc.Login(ctx, application.LoginInput{
		TenantCode: "demo",
		Email:      "admin@demo.local",
		Password:   "Admin123!",
		DeviceID:   "test-device",
		Platform:   "test",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("expected tokens")
	}
	if res.User.Email != "admin@demo.local" {
		t.Fatalf("unexpected user: %+v", res.User)
	}
	if len(res.User.Permissions) == 0 {
		t.Fatal("expected permissions")
	}

	refreshed, err := authSvc.Refresh(ctx, res.RefreshToken, "test-device")
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.AccessToken == "" {
		t.Fatal("expected new access token")
	}

	// old refresh should be invalid after rotation
	time.Sleep(10 * time.Millisecond)
	if _, err := authSvc.Refresh(ctx, res.RefreshToken, "test-device"); err == nil {
		t.Fatal("expected old refresh token to fail")
	}

	if err := authSvc.Logout(ctx, refreshed.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
}

func TestRegisterDevice(t *testing.T) {
	db := setupTestDB(t)
	tokenSvc := auth.NewTokenService(config.AuthConfig{
		AccessSecret:     "test-access-secret-32-characters!!",
		RefreshSecret:    "test-refresh-secret-32-characters!",
		AccessTTLMinutes: 15,
		RefreshTTLDays:   30,
		Issuer:           "sfa-test",
	})
	authSvc := application.NewAuthService(
		identitypersist.NewUserRepo(db),
		identitypersist.NewRefreshTokenRepo(db),
		identitypersist.NewDeviceRepo(db),
		tenantpersist.NewTenantRepo(db),
		tokenSvc,
	)
	ctx := context.Background()
	res, err := authSvc.Login(ctx, application.LoginInput{
		TenantCode: "demo", Email: "admin@demo.local", Password: "Admin123!",
		DeviceID: "dev-1", Platform: "web",
	})
	if err != nil {
		t.Fatal(err)
	}
	token := "push-token-abc"
	dev, err := authSvc.RegisterDevice(ctx, res.User.TenantID, res.User.ID, application.RegisterDeviceInput{
		DeviceID: "dev-1", Platform: "web", PushToken: &token,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dev.PushToken == nil || *dev.PushToken != token {
		t.Fatalf("unexpected device %+v", dev)
	}
	list, err := authSvc.ListDevices(ctx, res.User.TenantID, res.User.ID)
	if err != nil || len(list) == 0 {
		t.Fatalf("list devices: err=%v len=%d", err, len(list))
	}
	if err := authSvc.UnregisterDevice(ctx, res.User.ID, "dev-1"); err != nil {
		t.Fatal(err)
	}
}
