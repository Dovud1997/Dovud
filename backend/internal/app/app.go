package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/gateway"
	analyticsapp "github.com/Dovud1997/Dovud/backend/internal/modules/analytics/application"
	analyticspersist "github.com/Dovud1997/Dovud/backend/internal/modules/analytics/infrastructure/persistence"
	analyticshttp "github.com/Dovud1997/Dovud/backend/internal/modules/analytics/interfaces/http"
	auditapp "github.com/Dovud1997/Dovud/backend/internal/modules/audit/application"
	auditpersist "github.com/Dovud1997/Dovud/backend/internal/modules/audit/infrastructure/persistence"
	audithttp "github.com/Dovud1997/Dovud/backend/internal/modules/audit/interfaces/http"
	catalogapp "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/application"
	catalogpersist "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/infrastructure/persistence"
	cataloghttp "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/interfaces/http"
	crmapp "github.com/Dovud1997/Dovud/backend/internal/modules/crm/application"
	crmpersist "github.com/Dovud1997/Dovud/backend/internal/modules/crm/infrastructure/persistence"
	crmhttp "github.com/Dovud1997/Dovud/backend/internal/modules/crm/interfaces/http"
	docsapp "github.com/Dovud1997/Dovud/backend/internal/modules/documents/application"
	docspersist "github.com/Dovud1997/Dovud/backend/internal/modules/documents/infrastructure/persistence"
	docshttp "github.com/Dovud1997/Dovud/backend/internal/modules/documents/interfaces/http"
	ffapp "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/application"
	ffpersist "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/infrastructure/persistence"
	ffhttp "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/interfaces/http"
	financeapp "github.com/Dovud1997/Dovud/backend/internal/modules/finance/application"
	financepersist "github.com/Dovud1997/Dovud/backend/internal/modules/finance/infrastructure/persistence"
	financehttp "github.com/Dovud1997/Dovud/backend/internal/modules/finance/interfaces/http"
	identityapp "github.com/Dovud1997/Dovud/backend/internal/modules/identity/application"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	identityhttp "github.com/Dovud1997/Dovud/backend/internal/modules/identity/interfaces/http"
	notifyapp "github.com/Dovud1997/Dovud/backend/internal/modules/notifications/application"
	notifypersist "github.com/Dovud1997/Dovud/backend/internal/modules/notifications/infrastructure/persistence"
	notifyhttp "github.com/Dovud1997/Dovud/backend/internal/modules/notifications/interfaces/http"
	ordersapp "github.com/Dovud1997/Dovud/backend/internal/modules/orders/application"
	orderspersist "github.com/Dovud1997/Dovud/backend/internal/modules/orders/infrastructure/persistence"
	ordershttp "github.com/Dovud1997/Dovud/backend/internal/modules/orders/interfaces/http"
	orgapp "github.com/Dovud1997/Dovud/backend/internal/modules/organization/application"
	orgpersist "github.com/Dovud1997/Dovud/backend/internal/modules/organization/infrastructure/persistence"
	orghttp "github.com/Dovud1997/Dovud/backend/internal/modules/organization/interfaces/http"
	portalapp "github.com/Dovud1997/Dovud/backend/internal/modules/portal/application"
	portalpersist "github.com/Dovud1997/Dovud/backend/internal/modules/portal/infrastructure/persistence"
	portalhttp "github.com/Dovud1997/Dovud/backend/internal/modules/portal/interfaces/http"
	returnsapp "github.com/Dovud1997/Dovud/backend/internal/modules/returns/application"
	returnspersist "github.com/Dovud1997/Dovud/backend/internal/modules/returns/infrastructure/persistence"
	returnshttp "github.com/Dovud1997/Dovud/backend/internal/modules/returns/interfaces/http"
	syncapp "github.com/Dovud1997/Dovud/backend/internal/modules/sync/application"
	syncapply "github.com/Dovud1997/Dovud/backend/internal/modules/sync/infrastructure/applicator"
	syncpersist "github.com/Dovud1997/Dovud/backend/internal/modules/sync/infrastructure/persistence"
	synchttp "github.com/Dovud1997/Dovud/backend/internal/modules/sync/interfaces/http"
	tenantapp "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/application"
	tenantpersist "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/infrastructure/persistence"
	tenanthttp "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/interfaces/http"
	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/crypto"
	"github.com/Dovud1997/Dovud/backend/internal/platform/database"
	"github.com/Dovud1997/Dovud/backend/internal/platform/logger"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/Dovud1997/Dovud/backend/internal/platform/redisx"
	"github.com/Dovud1997/Dovud/backend/internal/platform/seed"
	"github.com/Dovud1997/Dovud/backend/internal/platform/storage"
	"github.com/Dovud1997/Dovud/backend/internal/platform/ws"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Application struct {
	Cfg    *config.Config
	Log    *slog.Logger
	DB     *gorm.DB
	Redis  *redisx.Client
	Store  storage.ObjectStore
	Router *fiber.App
}

func New(cfgPath string) (*Application, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	log := logger.New(cfg.App.Env)

	db, err := database.Connect(cfg.Database.DSN, cfg.App.Env)
	if err != nil {
		return nil, err
	}

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	if err := seed.Run(context.Background(), db, log); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	var redisClient *redisx.Client
	if cfg.Redis.Addr != "" {
		rc, rerr := redisx.Connect(cfg.Redis)
		if rerr != nil {
			log.Warn("redis unavailable", "error", rerr)
		} else {
			redisClient = rc
			log.Info("redis connected", "addr", cfg.Redis.Addr)
		}
	}

	objectStore, err := storage.Open(cfg.Minio, cfg.App.PublicBaseURL, cfg.Auth.AccessSecret, log)
	if err != nil {
		return nil, fmt.Errorf("storage: %w", err)
	}

	tokenSvc := auth.NewTokenService(cfg.Auth)
	outboxStore := outbox.NewStore(db)

	userRepo := identitypersist.NewUserRepo(db)
	roleRepo := identitypersist.NewRoleRepo(db)
	refreshRepo := identitypersist.NewRefreshTokenRepo(db)
	deviceRepo := identitypersist.NewDeviceRepo(db)

	tenantRepo := tenantpersist.NewTenantRepo(db)
	brandingRepo := tenantpersist.NewBrandingRepo(db)
	domainRepo := tenantpersist.NewDomainRepo(db)
	providerRepo := tenantpersist.NewProviderRepo(db)
	secretBox, err := crypto.NewSecretBox(cfg.Auth.AccessSecret)
	if err != nil {
		return nil, fmt.Errorf("provider crypto: %w", err)
	}

	companyRepo := orgpersist.NewCompanyRepo(db)
	branchRepo := orgpersist.NewBranchRepo(db)
	warehouseRepo := orgpersist.NewWarehouseRepo(db)

	manufacturerRepo := catalogpersist.NewManufacturerRepo(db)
	categoryRepo := catalogpersist.NewCategoryRepo(db)
	productRepo := catalogpersist.NewProductRepo(db)
	priceRepo := catalogpersist.NewPriceRepo(db)
	promotionRepo := catalogpersist.NewPromotionRepo(db)

	customerRepo := crmpersist.NewCustomerRepo(db)
	contactRepo := crmpersist.NewCustomerContactRepo(db)
	addressRepo := crmpersist.NewCustomerAddressRepo(db)
	customerCategoryRepo := crmpersist.NewCustomerCategoryRepo(db)

	agentRepo := ffpersist.NewAgentRepo(db)
	routeRepo := ffpersist.NewRouteRepo(db)
	visitRepo := ffpersist.NewVisitRepo(db)
	gpsRepo := ffpersist.NewGpsRepo(db)

	orderRepo := orderspersist.NewOrderRepo(db)
	returnRepo := returnspersist.NewReturnRepo(db)
	receivableRepo := financepersist.NewReceivableRepo(db)
	creditLimitRepo := financepersist.NewCreditLimitRepo(db)

	syncDeviceRepo := syncpersist.NewDeviceRepo(db)
	syncChangeRepo := syncpersist.NewChangeLogRepo(db)
	syncConflictRepo := syncpersist.NewConflictRepo(db)

	notifyRepo := notifypersist.NewNotificationRepo(db)
	kpiRepo := analyticspersist.NewKpiRepo(db)
	fileRepo := docspersist.NewFileRepo(db)
	documentRepo := docspersist.NewDocumentRepo(db)
	auditRepo := auditpersist.NewAuditRepo(db)
	customerUserRepo := portalpersist.NewCustomerUserRepo(db)

	lockCfg := auth.LockoutConfig{
		MaxAttempts: cfg.Security.LoginMaxAttempts,
		Window:      time.Duration(cfg.Security.LoginWindowMinutes) * time.Minute,
		LockFor:     time.Duration(cfg.Security.LoginLockoutMinutes) * time.Minute,
	}
	var loginGuard auth.LoginGuard = auth.NewMemoryLoginGuard(lockCfg)
	if redisClient != nil {
		loginGuard = auth.NewRedisLoginGuard(redisClient, lockCfg)
	}
	authSvc := identityapp.NewAuthService(userRepo, refreshRepo, deviceRepo, tenantRepo, tokenSvc).WithLoginGuard(loginGuard)
	rbacSvc := identityapp.NewRBACService(userRepo, roleRepo)
	docsSvc := docsapp.NewService(fileRepo, documentRepo, objectStore, outboxStore)
	tenantSvc := tenantapp.NewTenantService(tenantRepo, brandingRepo, domainRepo).
		WithProviders(providerRepo, secretBox).
		WithAssets(docsSvc)
	orgSvc := orgapp.NewService(companyRepo, branchRepo, warehouseRepo)
	catalogSvc := catalogapp.NewService(manufacturerRepo, categoryRepo, productRepo, priceRepo, promotionRepo)
	crmSvc := crmapp.NewService(customerRepo, contactRepo, addressRepo, customerCategoryRepo)
	ffSvc := ffapp.NewService(agentRepo, routeRepo, visitRepo, gpsRepo)
	ordersSvc := ordersapp.NewService(orderRepo)
	returnsSvc := returnsapp.NewService(returnRepo)
	financeSvc := financeapp.NewService(receivableRepo, creditLimitRepo)
	syncSvc := syncapp.NewService(syncDeviceRepo, syncChangeRepo, syncConflictRepo)
	if redisClient != nil {
		syncSvc.WithLocker(redisClient)
	}
	liveHub := ws.NewHub(log)
	if redisClient != nil {
		liveHub.WithRedis(redisClient.Raw())
	}
	syncSvc.WithLive(liveHub)
	syncSvc.WithApplicator(syncapply.New(customerRepo, orderRepo, visitRepo, returnRepo))
	crmSvc.WithSync(syncSvc)
	ffSvc.WithSync(syncSvc)
	ordersSvc.WithSync(syncSvc)
	catalogSvc.WithSync(syncSvc)
	returnsSvc.WithSync(syncSvc)
	notifySvc := notifyapp.NewService(notifyRepo, outboxStore).WithLive(liveHub)
	analyticsSvc := analyticsapp.NewService(kpiRepo, db)
	auditSvc := auditapp.NewService(auditRepo, outboxStore)
	auditWriter := audithttp.NewHTTPWriter(auditSvc, log)
	portalSvc := portalapp.NewService(customerUserRepo, customerRepo, orderRepo, receivableRepo, documentRepo, userRepo)

	router := gateway.NewRouter(gateway.Deps{
		Config:        cfg,
		TokenService:  tokenSvc,
		AuditWriter:   auditWriter,
		Identity:      identityhttp.NewHandler(authSvc, rbacSvc),
		Tenant:        tenanthttp.NewHandler(tenantSvc),
		Organization:  orghttp.NewHandler(orgSvc),
		Catalog:       cataloghttp.NewHandler(catalogSvc),
		CRM:           crmhttp.NewHandler(crmSvc),
		FieldForce:    ffhttp.NewHandler(ffSvc),
		Orders:        ordershttp.NewHandler(ordersSvc),
		Returns:       returnshttp.NewHandler(returnsSvc),
		Finance:       financehttp.NewHandler(financeSvc),
		Sync:          synchttp.NewHandler(syncSvc),
		Notifications: notifyhttp.NewHandler(notifySvc),
		Analytics:     analyticshttp.NewHandler(analyticsSvc),
		Documents:     docshttp.NewHandler(docsSvc, objectStore),
		Audit:         audithttp.NewHandler(auditSvc),
		Portal:        portalhttp.NewHandler(portalSvc),
	})
	ws.Register(router, liveHub, tokenSvc)

	return &Application{Cfg: cfg, Log: log, DB: db, Redis: redisClient, Store: objectStore, Router: router}, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&tenantpersist.TenantModel{},
		&tenantpersist.BrandingModel{},
		&tenantpersist.DomainModel{},
		&tenantpersist.ProviderModel{},
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
	)
}

func (a *Application) Run() error {
	a.Log.Info("starting api", "addr", a.Cfg.App.HTTPAddr, "env", a.Cfg.App.Env, "storage", a.Store.Driver())
	return a.Router.Listen(a.Cfg.App.HTTPAddr)
}
