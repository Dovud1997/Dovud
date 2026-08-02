package gateway

import (
	analytichttp "github.com/Dovud1997/Dovud/backend/internal/modules/analytics/interfaces/http"
	audithttp "github.com/Dovud1997/Dovud/backend/internal/modules/audit/interfaces/http"
	cataloghttp "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/interfaces/http"
	crmhttp "github.com/Dovud1997/Dovud/backend/internal/modules/crm/interfaces/http"
	docshttp "github.com/Dovud1997/Dovud/backend/internal/modules/documents/interfaces/http"
	ffhttp "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/interfaces/http"
	financehttp "github.com/Dovud1997/Dovud/backend/internal/modules/finance/interfaces/http"
	identityhttp "github.com/Dovud1997/Dovud/backend/internal/modules/identity/interfaces/http"
	notifyhttp "github.com/Dovud1997/Dovud/backend/internal/modules/notifications/interfaces/http"
	ordershttp "github.com/Dovud1997/Dovud/backend/internal/modules/orders/interfaces/http"
	orghttp "github.com/Dovud1997/Dovud/backend/internal/modules/organization/interfaces/http"
	portalhttp "github.com/Dovud1997/Dovud/backend/internal/modules/portal/interfaces/http"
	returnshttp "github.com/Dovud1997/Dovud/backend/internal/modules/returns/interfaces/http"
	synchttp "github.com/Dovud1997/Dovud/backend/internal/modules/sync/interfaces/http"
	tenanthttp "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/interfaces/http"
	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Deps struct {
	TokenService  *auth.TokenService
	AuditWriter   httpx.AuditWriter
	Identity      *identityhttp.Handler
	Tenant        *tenanthttp.Handler
	Organization  *orghttp.Handler
	Catalog       *cataloghttp.Handler
	CRM           *crmhttp.Handler
	FieldForce    *ffhttp.Handler
	Orders        *ordershttp.Handler
	Returns       *returnshttp.Handler
	Finance       *financehttp.Handler
	Sync          *synchttp.Handler
	Notifications *notifyhttp.Handler
	Analytics     *analytichttp.Handler
	Documents     *docshttp.Handler
	Audit         *audithttp.Handler
	Portal        *portalhttp.Handler
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
	if deps.Documents != nil {
		deps.Documents.RegisterPublic(v1)
	}

	protected := v1.Group("", httpx.AuthMiddleware(deps.TokenService))
	if deps.AuditWriter != nil {
		protected.Use(httpx.AuditMiddleware(deps.AuditWriter))
	}
	deps.Identity.RegisterProtected(protected)
	deps.Tenant.RegisterProtected(protected)
	deps.Organization.Register(protected)
	deps.Catalog.Register(protected)
	deps.CRM.Register(protected)
	deps.FieldForce.Register(protected)
	deps.Orders.Register(protected)
	deps.Returns.Register(protected)
	deps.Finance.Register(protected)
	deps.Sync.Register(protected)
	deps.Notifications.Register(protected)
	deps.Analytics.Register(protected)
	if deps.Documents != nil {
		deps.Documents.Register(protected)
	}
	if deps.Audit != nil {
		deps.Audit.Register(protected)
	}
	if deps.Portal != nil {
		deps.Portal.Register(protected)
	}

	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	return httpx.Fail(c, err)
}
