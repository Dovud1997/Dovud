package seed

import (
	"context"
	"log/slog"
	"strings"

	catalogdomain "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/domain"
	catalogpersist "github.com/Dovud1997/Dovud/backend/internal/modules/catalog/infrastructure/persistence"
	crmdomain "github.com/Dovud1997/Dovud/backend/internal/modules/crm/domain"
	crmpersist "github.com/Dovud1997/Dovud/backend/internal/modules/crm/infrastructure/persistence"
	identitydomain "github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	orgdomain "github.com/Dovud1997/Dovud/backend/internal/modules/organization/domain"
	orgpersist "github.com/Dovud1997/Dovud/backend/internal/modules/organization/infrastructure/persistence"
	tenantdomain "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/domain"
	tenantpersist "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var permissionCodes = []string{
	"users:read", "users:write",
	"roles:read", "roles:write",
	"tenant:read", "tenant:write",
	"branches:read", "branches:write",
	"warehouses:read", "warehouses:write",
	"catalog:read", "catalog:write",
	"customers:read", "customers:write",
	"agents:read", "agents:write",
	"routes:read", "routes:write",
	"visits:read", "visits:write",
	"orders:read", "orders:write", "orders:approve",
	"returns:read", "returns:write",
	"finance:read", "finance:write",
	"documents:read", "documents:write",
	"notifications:read", "notifications:write",
	"analytics:read",
	"audit:read",
	"sync:use",
}

func Run(ctx context.Context, db *gorm.DB, log *slog.Logger) error {
	if log == nil {
		log = slog.Default()
	}
	if err := ensurePermissions(ctx, db); err != nil {
		return err
	}

	tenantRepo := tenantpersist.NewTenantRepo(db)
	brandingRepo := tenantpersist.NewBrandingRepo(db)
	domainRepo := tenantpersist.NewDomainRepo(db)
	userRepo := identitypersist.NewUserRepo(db)
	roleRepo := identitypersist.NewRoleRepo(db)

	tenant, err := tenantRepo.FindByCode(ctx, "demo")
	if err != nil {
		tenant = &tenantdomain.Tenant{
			Code: "demo", Name: "Demo Company", Status: "active",
			DefaultLocale: "ru", DefaultCurrency: "UZS", Timezone: "Asia/Tashkent",
		}
		if err := tenantRepo.Create(ctx, tenant); err != nil {
			return err
		}
		log.Info("seeded tenant", "code", tenant.Code, "id", tenant.ID)
	}

	if _, err := brandingRepo.GetByTenantID(ctx, tenant.ID); err != nil {
		b := &tenantdomain.Branding{
			TenantID: tenant.ID, AppName: "SFA Demo",
			PrimaryColor: "#0F766E", SecondaryColor: "#134E4A", AccentColor: "#F59E0B",
			ThemeModeDefault: "light",
		}
		if err := brandingRepo.Upsert(ctx, b); err != nil {
			return err
		}
	}

	domains, _ := domainRepo.List(ctx, tenant.ID)
	if len(domains) == 0 {
		_ = domainRepo.Create(ctx, &tenantdomain.Domain{TenantID: tenant.ID, Host: "localhost", IsPrimary: true})
	}

	ownerRole, err := ensureRole(ctx, roleRepo, &tenant.ID, "tenant_owner", "Tenant Owner", true)
	if err != nil {
		return err
	}
	agentRole, err := ensureRole(ctx, roleRepo, &tenant.ID, "sales_agent", "Sales Agent", true)
	if err != nil {
		return err
	}

	allPermIDs, err := roleRepo.PermissionIDsByCodes(ctx, permissionCodes)
	if err != nil {
		return err
	}
	if err := roleRepo.SetPermissions(ctx, ownerRole.ID, allPermIDs); err != nil {
		return err
	}
	agentPerms, err := roleRepo.PermissionIDsByCodes(ctx, []string{
		"customers:read", "catalog:read", "orders:read", "orders:write",
		"visits:read", "visits:write", "routes:read", "sync:use",
		"notifications:read", "documents:read", "documents:write", "finance:read",
	})
	if err != nil {
		return err
	}
	if err := roleRepo.SetPermissions(ctx, agentRole.ID, agentPerms); err != nil {
		return err
	}

	if _, err := userRepo.FindByEmail(ctx, tenant.ID, "admin@demo.local"); err != nil {
		hash, err := auth.HashPassword("Admin123!")
		if err != nil {
			return err
		}
		admin := &identitydomain.User{
			TenantID: tenant.ID, Email: "admin@demo.local", PasswordHash: hash,
			FullName: "Demo Admin", Status: "active", Locale: "ru", ThemePreference: "system", Version: 1,
		}
		if err := userRepo.Create(ctx, admin); err != nil {
			return err
		}
		if err := userRepo.ReplaceRoles(ctx, admin.ID, []uuid.UUID{ownerRole.ID}); err != nil {
			return err
		}
		log.Info("seeded admin user", "email", admin.Email, "password", "Admin123!")
	}

	if _, err := userRepo.FindByEmail(ctx, tenant.ID, "agent@demo.local"); err != nil {
		hash, err := auth.HashPassword("Agent123!")
		if err != nil {
			return err
		}
		agent := &identitydomain.User{
			TenantID: tenant.ID, Email: "agent@demo.local", PasswordHash: hash,
			FullName: "Demo Agent", Status: "active", Locale: "ru", ThemePreference: "system", Version: 1,
		}
		if err := userRepo.Create(ctx, agent); err != nil {
			return err
		}
		if err := userRepo.ReplaceRoles(ctx, agent.ID, []uuid.UUID{agentRole.ID}); err != nil {
			return err
		}
		log.Info("seeded agent user", "email", agent.Email, "password", "Agent123!")
	}

	if err := seedBusinessDemo(ctx, db, tenant.ID, log); err != nil {
		return err
	}

	return nil
}

func seedBusinessDemo(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, log *slog.Logger) error {
	companyRepo := orgpersist.NewCompanyRepo(db)
	branchRepo := orgpersist.NewBranchRepo(db)
	warehouseRepo := orgpersist.NewWarehouseRepo(db)
	manufacturerRepo := catalogpersist.NewManufacturerRepo(db)
	categoryRepo := catalogpersist.NewCategoryRepo(db)
	productRepo := catalogpersist.NewProductRepo(db)
	priceRepo := catalogpersist.NewPriceRepo(db)
	customerRepo := crmpersist.NewCustomerRepo(db)
	contactRepo := crmpersist.NewCustomerContactRepo(db)

	companies, _, err := companyRepo.List(ctx, tenantID, 1, 1)
	if err != nil {
		return err
	}
	var companyID uuid.UUID
	if len(companies) == 0 {
		c := &orgdomain.Company{TenantID: tenantID, Code: "HQ", Name: "Demo HQ", Status: "active"}
		if err := companyRepo.Create(ctx, c); err != nil {
			return err
		}
		companyID = c.ID
		log.Info("seeded company", "code", c.Code)
	} else {
		companyID = companies[0].ID
	}

	branches, _, err := branchRepo.List(ctx, tenantID, 1, 1)
	if err != nil {
		return err
	}
	var branchID uuid.UUID
	if len(branches) == 0 {
		addr := "Tashkent, Demo Street 1"
		lat, lng := 41.3111, 69.2797
		b := &orgdomain.Branch{TenantID: tenantID, CompanyID: &companyID, Code: "TAS-01", Name: "Tashkent Branch", Address: &addr, Lat: &lat, Lng: &lng, Status: "active"}
		if err := branchRepo.Create(ctx, b); err != nil {
			return err
		}
		branchID = b.ID
		w := &orgdomain.Warehouse{TenantID: tenantID, BranchID: branchID, Code: "WH-01", Name: "Main Warehouse", Type: "main", Status: "active"}
		if err := warehouseRepo.Create(ctx, w); err != nil {
			return err
		}
		log.Info("seeded branch and warehouse", "branch", b.Code)
	} else {
		branchID = branches[0].ID
	}

	manufacturers, _, err := manufacturerRepo.List(ctx, tenantID, 1, 1)
	if err != nil {
		return err
	}
	var manufacturerID uuid.UUID
	if len(manufacturers) == 0 {
		m := &catalogdomain.Manufacturer{TenantID: tenantID, Code: "ACME", Name: "Acme Foods", Status: "active"}
		if err := manufacturerRepo.Create(ctx, m); err != nil {
			return err
		}
		manufacturerID = m.ID
	} else {
		manufacturerID = manufacturers[0].ID
	}

	categories, err := categoryRepo.List(ctx, tenantID)
	if err != nil {
		return err
	}
	var categoryID uuid.UUID
	if len(categories) == 0 {
		cat := &catalogdomain.Category{TenantID: tenantID, Code: "BEV", Name: "Beverages", SortOrder: 1}
		if err := categoryRepo.Create(ctx, cat); err != nil {
			return err
		}
		categoryID = cat.ID
	} else {
		categoryID = categories[0].ID
	}

	products, _, err := productRepo.List(ctx, tenantID, "", 1, 1)
	if err != nil {
		return err
	}
	var productID uuid.UUID
	if len(products) == 0 {
		p := &catalogdomain.Product{
			TenantID: tenantID, SKU: "SKU-1001", Name: "Demo Cola 0.5L",
			CategoryID: &categoryID, ManufacturerID: &manufacturerID, Unit: "pcs", VATRate: 12, IsActive: true,
		}
		if err := productRepo.Create(ctx, p); err != nil {
			return err
		}
		productID = p.ID
		pl := &catalogdomain.PriceList{TenantID: tenantID, Code: "STD", Name: "Standard", Currency: "UZS", IsDefault: true}
		if err := priceRepo.CreatePriceList(ctx, pl); err != nil {
			return err
		}
		if err := priceRepo.UpsertPrice(ctx, &catalogdomain.ProductPrice{
			TenantID: tenantID, PriceListID: pl.ID, ProductID: productID, Amount: 8500, Currency: "UZS",
		}); err != nil {
			return err
		}
		log.Info("seeded catalog", "sku", p.SKU)
	}

	customers, _, err := customerRepo.List(ctx, tenantID, 1, 1)
	if err != nil {
		return err
	}
	if len(customers) == 0 {
		addr := "Chilonzor, Market 12"
		lat, lng := 41.2856, 69.2034
		cust := &crmdomain.Customer{
			TenantID: tenantID, BranchID: &branchID, Code: "C-100", Name: "Demo Market",
			Type: "outlet", Status: "active", CreditLimit: 5000000, Address: &addr, Lat: &lat, Lng: &lng,
		}
		if err := customerRepo.Create(ctx, cust); err != nil {
			return err
		}
		_ = contactRepo.Create(ctx, &crmdomain.CustomerContact{
			CustomerID: cust.ID, FullName: "Dilshod Karimov", Phone: "+998901112233", IsPrimary: true,
		})
		log.Info("seeded customer", "code", cust.Code)
	}
	return nil
}

func ensurePermissions(ctx context.Context, db *gorm.DB) error {
	for _, code := range permissionCodes {
		var count int64
		if err := db.WithContext(ctx).Model(&identitypersist.PermissionModel{}).Where("code = ?", code).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		p := identitypersist.PermissionModel{
			ID: uuid.New(), Code: code, Description: strings.ReplaceAll(code, ":", " "),
		}
		if err := db.WithContext(ctx).Create(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureRole(ctx context.Context, roles identitydomain.RoleRepository, tenantID *uuid.UUID, code, name string, system bool) (*identitydomain.Role, error) {
	role, err := roles.FindByCode(ctx, tenantID, code)
	if err == nil {
		return role, nil
	}
	role = &identitydomain.Role{TenantID: tenantID, Code: code, Name: name, IsSystem: system}
	if err := roles.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}
