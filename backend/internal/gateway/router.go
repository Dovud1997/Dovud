package gateway

import (
	cataloghttp "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/interfaces/http"
	crmhttp "github.com/Dovud1997/Dovud/backend/internal/modules/crm/interfaces/http"
	identityhttp "github.com/Dovud1997/Dovud/backend/internal/modules/identity/interfaces/http"
	orghttp "github.com/Dovud1997/Dovud/backend/internal/modules/organization/interfaces/http"
	tenanthttp "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/interfaces/http"
	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Deps struct {
	TokenService *auth.TokenService
	Identity     *identityhttp.Handler
	Tenant       *tenanthttp.Handler
	Organization *orghttp.Handler
	Catalog      *cataloghttp.Handler
	CRM          *crmhttp.Handler
}

func NewRouter(deps Deps) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "SFA API",
		ErrorHandler: errorHandler,
	})

	app.Use(recover.New())
	app.Use(httpx.RequestIDMiddleware())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Tenant-ID, X-Request-ID, Idempotency-Key, Accept-Language",
	}))

	v1 := app.Group("/api/v1")
	deps.Tenant.RegisterPublic(v1)
	deps.Identity.RegisterPublic(v1)

	protected := v1.Group("", httpx.AuthMiddleware(deps.TokenService))
	deps.Identity.RegisterProtected(protected)
	deps.Tenant.RegisterProtected(protected)
	deps.Organization.Register(protected)
	deps.Catalog.Register(protected)
	deps.CRM.Register(protected)

	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	return httpx.Fail(c, err)
}
