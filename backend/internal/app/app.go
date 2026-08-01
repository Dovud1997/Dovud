package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Dovud1997/Dovud/backend/internal/gateway"
	identityapp "github.com/Dovud1997/Dovud/backend/internal/modules/identity/application"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	identityhttp "github.com/Dovud1997/Dovud/backend/internal/modules/identity/interfaces/http"
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

	authSvc := identityapp.NewAuthService(userRepo, refreshRepo, deviceRepo, tenantRepo, tokenSvc)
	rbacSvc := identityapp.NewRBACService(userRepo, roleRepo)
	tenantSvc := tenantapp.NewTenantService(tenantRepo, brandingRepo, domainRepo)

	identityHandler := identityhttp.NewHandler(authSvc, rbacSvc)
	tenantHandler := tenanthttp.NewHandler(tenantSvc)

	router := gateway.NewRouter(gateway.Deps{
		TokenService: tokenSvc,
		Identity:     identityHandler,
		Tenant:       tenantHandler,
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
	)
}

func (a *Application) Run() error {
	a.Log.Info("starting api", "addr", a.Cfg.App.HTTPAddr, "env", a.Cfg.App.Env)
	return a.Router.Listen(a.Cfg.App.HTTPAddr)
}
