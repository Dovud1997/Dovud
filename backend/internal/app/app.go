package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Dovud1997/Dovud/backend/internal/gateway"
	catalogapp "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/application"
	catalogpersist "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/infrastructure/persistence"
	cataloghttp "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/interfaces/http"
	crmapp "github.com/Dovud1997/Dovud/backend/internal/modules/crm/application"
	crmpersist "github.com/Dovud1997/Dovud/backend/internal/modules/crm/infrastructure/persistence"
	crmhttp "github.com/Dovud1997/Dovud/backend/internal/modules/crm/interfaces/http"
	ffapp "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/application"
	ffpersist "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/infrastructure/persistence"
	ffhttp "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/interfaces/http"
	identityapp "github.com/Dovud1997/Dovud/backend/internal/modules/identity/application"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	identityhttp "github.com/Dovud1997/Dovud/backend/internal/modules/identity/interfaces/http"
	ordersapp "github.com/Dovud1997/Dovud/backend/internal/modules/orders/application"
	orderspersist "github.com/Dovud1997/Dovud/backend/internal/modules/orders/infrastructure/persistence"
	ordershttp "github.com/Dovud1997/Dovud/backend/internal/modules/orders/interfaces/http"
	orgapp "github.com/Dovud1997/Dovud/backend/internal/modules/organization/application"
	orgpersist "github.com/Dovud1997/Dovud/backend/internal/modules/organization/infrastructure/persistence"
	orghttp "github.com/Dovud1997/Dovud/backend/internal/modules/organization/interfaces/http"
	tenantapp "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/application"
	tenantpersist "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/infrastructure/persistence"
	tenanthttp "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/interfaces/http"
	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/database"
	"github.com/Dovud1997/Dovud/backend/internal/platform/logger"
	"github.com/Dovud1997/Dovud/backend/internal/platform/seed"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Application struct {
	Cfg    *config.Config
	Log    *slog.Logger
	DB     *gorm.DB
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

	tokenSvc := auth.NewTokenService(cfg.Auth)

	userRepo := identitypersist.NewUserRepo(db)
	roleRepo := identitypersist.NewRoleRepo(db)
	refreshRepo := identitypersist.NewRefreshTokenRepo(db)
	deviceRepo := identitypersist.NewDeviceRepo(db)

	tenantRepo := tenantpersist.NewTenantRepo(db)
	brandingRepo := tenantpersist.NewBrandingRepo(db)
	domainRepo := tenantpersist.NewDomainRepo(db)

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

	authSvc := identityapp.NewAuthService(userRepo, refreshRepo, deviceRepo, tenantRepo, tokenSvc)
	rbacSvc := identityapp.NewRBACService(userRepo, roleRepo)
	tenantSvc := tenantapp.NewTenantService(tenantRepo, brandingRepo, domainRepo)
	orgSvc := orgapp.NewService(companyRepo, branchRepo, warehouseRepo)
	catalogSvc := catalogapp.NewService(manufacturerRepo, categoryRepo, productRepo, priceRepo, promotionRepo)
	crmSvc := crmapp.NewService(customerRepo, contactRepo, addressRepo, customerCategoryRepo)
	ffSvc := ffapp.NewService(agentRepo, routeRepo, visitRepo, gpsRepo)
	ordersSvc := ordersapp.NewService(orderRepo)

	router := gateway.NewRouter(gateway.Deps{
		TokenService: tokenSvc,
		Identity:     identityhttp.NewHandler(authSvc, rbacSvc),
		Tenant:       tenanthttp.NewHandler(tenantSvc),
		Organization: orghttp.NewHandler(orgSvc),
		Catalog:      cataloghttp.NewHandler(catalogSvc),
		CRM:          crmhttp.NewHandler(crmSvc),
		FieldForce:   ffhttp.NewHandler(ffSvc),
		Orders:       ordershttp.NewHandler(ordersSvc),
	})

	return &Application{Cfg: cfg, Log: log, DB: db, Router: router}, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
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
	)
}

func (a *Application) Run() error {
	a.Log.Info("starting api", "addr", a.Cfg.App.HTTPAddr, "env", a.Cfg.App.Env)
	return a.Router.Listen(a.Cfg.App.HTTPAddr)
}
