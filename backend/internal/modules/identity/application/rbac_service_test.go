package application_test

import (
	"context"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/application"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	tenantpersist "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/infrastructure/persistence"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestRBACListRolesIncludesPermissionCodes(t *testing.T) {
	db := setupTestDB(t)
	users := identitypersist.NewUserRepo(db)
	roles := identitypersist.NewRoleRepo(db)
	svc := application.NewRBACService(users, roles)
	ctx := context.Background()
	tenantID := demoTenantID(t, db)

	list, err := svc.ListRoles(ctx, tenantID)
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected seeded roles")
	}
	var owner *application.RoleDTO
	for i := range list {
		if list[i].Code == "tenant_owner" {
			owner = &list[i]
			break
		}
	}
	if owner == nil {
		t.Fatal("expected tenant_owner role")
	}
	if len(owner.PermissionCodes) == 0 {
		t.Fatal("expected permission_codes on system role")
	}
}

func TestRBACUpdateUserAndLockSystemRole(t *testing.T) {
	db := setupTestDB(t)
	users := identitypersist.NewUserRepo(db)
	roles := identitypersist.NewRoleRepo(db)
	svc := application.NewRBACService(users, roles)
	ctx := context.Background()
	tenantID := demoTenantID(t, db)

	listed, _, err := svc.ListUsers(ctx, tenantID, 1, 50)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	var agent *application.UserDTO
	for i := range listed {
		if listed[i].Email == "agent@demo.local" {
			agent = &listed[i]
			break
		}
	}
	if agent == nil {
		t.Fatal("expected agent user")
	}
	if len(agent.RoleIDs) == 0 {
		t.Fatal("expected role_ids on user DTO")
	}

	name := "Agent Updated"
	status := "disabled"
	updated, err := svc.UpdateUser(ctx, tenantID, agent.ID, application.UpdateUserInput{
		FullName: &name, Status: &status,
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.FullName != name || updated.Status != "disabled" {
		t.Fatalf("unexpected update: %+v", updated)
	}

	roleList, err := svc.ListRoles(ctx, tenantID)
	if err != nil {
		t.Fatal(err)
	}
	var systemID uuid.UUID
	for _, r := range roleList {
		if r.IsSystem {
			systemID = r.ID
			break
		}
	}
	if systemID == uuid.Nil {
		t.Fatal("expected system role")
	}
	if err := svc.SetRolePermissions(ctx, tenantID, systemID, []string{"users:read"}); err == nil {
		t.Fatal("expected system role lock")
	}

	created, err := svc.CreateRole(ctx, tenantID, application.CreateRoleInput{
		Code: "field_lead", Name: "Field Lead", PermissionCodes: []string{"routes:read", "orders:read"},
	})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if len(created.PermissionCodes) != 2 {
		t.Fatalf("expected 2 permission codes, got %v", created.PermissionCodes)
	}
	if err := svc.SetRolePermissions(ctx, tenantID, created.ID, []string{"orders:write"}); err != nil {
		t.Fatalf("set custom role perms: %v", err)
	}
}

func demoTenantID(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	tenant, err := tenantpersist.NewTenantRepo(db).FindByCode(context.Background(), "demo")
	if err != nil {
		t.Fatalf("demo tenant: %v", err)
	}
	return tenant.ID
}
