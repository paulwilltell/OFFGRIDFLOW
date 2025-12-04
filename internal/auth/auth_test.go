package auth

import (
	"context"
	"testing"
)

func TestRBACAuthorizer_Authorize(t *testing.T) {
	auth := NewRBACAuthorizer()
	ctx := context.Background()

	tests := []struct {
		name        string
		user        User
		action      string
		resource    string
		wantAllowed bool
	}{
		{
			name:        "admin has full access",
			user:        User{ID: "1", Email: "admin@test.com", Role: "admin"},
			action:      ActionDelete,
			resource:    ResourceUsers,
			wantAllowed: true,
		},
		{
			name:        "editor can read emissions",
			user:        User{ID: "2", Email: "editor@test.com", Role: "editor"},
			action:      ActionRead,
			resource:    ResourceEmissions,
			wantAllowed: true,
		},
		{
			name:        "editor can write emissions",
			user:        User{ID: "2", Email: "editor@test.com", Role: "editor"},
			action:      ActionWrite,
			resource:    ResourceEmissions,
			wantAllowed: true,
		},
		{
			name:        "editor cannot delete",
			user:        User{ID: "2", Email: "editor@test.com", Role: "editor"},
			action:      ActionDelete,
			resource:    ResourceEmissions,
			wantAllowed: false,
		},
		{
			name:        "viewer can read emissions",
			user:        User{ID: "3", Email: "viewer@test.com", Role: "viewer"},
			action:      ActionRead,
			resource:    ResourceEmissions,
			wantAllowed: true,
		},
		{
			name:        "viewer cannot write",
			user:        User{ID: "3", Email: "viewer@test.com", Role: "viewer"},
			action:      ActionWrite,
			resource:    ResourceEmissions,
			wantAllowed: false,
		},
		{
			name:        "unknown role has no access",
			user:        User{ID: "4", Email: "unknown@test.com", Role: "unknown"},
			action:      ActionRead,
			resource:    ResourceEmissions,
			wantAllowed: false,
		},
		{
			name:        "user with multiple roles - admin role grants access",
			user:        User{ID: "5", Email: "multi@test.com", Role: "viewer", Roles: []string{"viewer", "admin"}},
			action:      ActionDelete,
			resource:    ResourceUsers,
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := auth.Authorize(ctx, tt.user, tt.action, tt.resource)
			if err != nil {
				t.Fatalf("Authorize() error = %v", err)
			}
			if allowed != tt.wantAllowed {
				t.Errorf("Authorize() = %v, want %v", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestRBACAuthorizer_AddRole(t *testing.T) {
	auth := NewRBACAuthorizer()
	ctx := context.Background()

	// Add custom role
	customRole := "auditor"
	permissions := []Permission{
		{Action: ActionRead, Resource: ResourceEmissions},
		{Action: ActionRead, Resource: ResourceCompliance},
	}
	auth.AddRole(customRole, permissions)

	user := User{ID: "1", Email: "auditor@test.com", Role: customRole}

	// Should have read access to emissions
	allowed, err := auth.Authorize(ctx, user, ActionRead, ResourceEmissions)
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if !allowed {
		t.Error("Expected auditor to have read access to emissions")
	}

	// Should not have write access
	allowed, err = auth.Authorize(ctx, user, ActionWrite, ResourceEmissions)
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if allowed {
		t.Error("Expected auditor to not have write access to emissions")
	}
}

func TestRBACAuthorizer_RemoveRole(t *testing.T) {
	auth := NewRBACAuthorizer()
	ctx := context.Background()

	// Remove viewer role
	auth.RemoveRole("viewer")

	user := User{ID: "1", Email: "viewer@test.com", Role: "viewer"}
	allowed, err := auth.Authorize(ctx, user, ActionRead, ResourceEmissions)
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if allowed {
		t.Error("Expected removed role to have no access")
	}
}

func TestRBACAuthorizer_GetRolePermissions(t *testing.T) {
	auth := NewRBACAuthorizer()

	perms := auth.GetRolePermissions("editor")
	if len(perms) == 0 {
		t.Error("Expected editor to have permissions")
	}

	// Verify we got a copy
	perms[0].Action = "modified"
	originalPerms := auth.GetRolePermissions("editor")
	if originalPerms[0].Action == "modified" {
		t.Error("GetRolePermissions should return a copy")
	}
}

func TestRBACAuthorizer_ListRoles(t *testing.T) {
	auth := NewRBACAuthorizer()
	roles := auth.ListRoles()

	expectedRoles := map[string]bool{
		"admin":  true,
		"editor": true,
		"viewer": true,
	}

	if len(roles) != len(expectedRoles) {
		t.Errorf("Expected %d roles, got %d", len(expectedRoles), len(roles))
	}

	for _, role := range roles {
		if !expectedRoles[role] {
			t.Errorf("Unexpected role: %s", role)
		}
	}
}

func TestAuthorizationHelpers(t *testing.T) {
	auth := NewRBACAuthorizer()
	ctx := context.Background()

	adminUser := User{ID: "1", Email: "admin@test.com", Role: "admin"}
	viewerUser := User{ID: "2", Email: "viewer@test.com", Role: "viewer"}

	// Test CanRead
	if !CanRead(ctx, auth, adminUser, ResourceEmissions) {
		t.Error("Admin should be able to read")
	}
	if !CanRead(ctx, auth, viewerUser, ResourceEmissions) {
		t.Error("Viewer should be able to read")
	}

	// Test CanWrite
	if !CanWrite(ctx, auth, adminUser, ResourceEmissions) {
		t.Error("Admin should be able to write")
	}
	if CanWrite(ctx, auth, viewerUser, ResourceEmissions) {
		t.Error("Viewer should not be able to write")
	}

	// Test CanDelete
	if !CanDelete(ctx, auth, adminUser, ResourceUsers) {
		t.Error("Admin should be able to delete")
	}
	if CanDelete(ctx, auth, viewerUser, ResourceUsers) {
		t.Error("Viewer should not be able to delete")
	}

	// Test CanAdmin
	if !CanAdmin(ctx, auth, adminUser, ResourceSettings) {
		t.Error("Admin should have admin permission")
	}
	if CanAdmin(ctx, auth, viewerUser, ResourceSettings) {
		t.Error("Viewer should not have admin permission")
	}
}

func TestWildcardPermissions(t *testing.T) {
	auth := NewRBACAuthorizerWithPermissions(map[string][]Permission{
		"superuser": {
			{Action: "*", Resource: "*"},
		},
		"emissions_admin": {
			{Action: "*", Resource: ResourceEmissions},
		},
		"global_reader": {
			{Action: ActionRead, Resource: "*"},
		},
	})

	ctx := context.Background()

	tests := []struct {
		name        string
		role        string
		action      string
		resource    string
		wantAllowed bool
	}{
		{
			name:        "superuser can do anything",
			role:        "superuser",
			action:      ActionDelete,
			resource:    ResourceUsers,
			wantAllowed: true,
		},
		{
			name:        "emissions_admin can delete emissions",
			role:        "emissions_admin",
			action:      ActionDelete,
			resource:    ResourceEmissions,
			wantAllowed: true,
		},
		{
			name:        "emissions_admin cannot delete users",
			role:        "emissions_admin",
			action:      ActionDelete,
			resource:    ResourceUsers,
			wantAllowed: false,
		},
		{
			name:        "global_reader can read anything",
			role:        "global_reader",
			action:      ActionRead,
			resource:    ResourceBilling,
			wantAllowed: true,
		},
		{
			name:        "global_reader cannot write",
			role:        "global_reader",
			action:      ActionWrite,
			resource:    ResourceBilling,
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{ID: "1", Email: "test@test.com", Role: tt.role}
			allowed, err := auth.Authorize(ctx, user, tt.action, tt.resource)
			if err != nil {
				t.Fatalf("Authorize() error = %v", err)
			}
			if allowed != tt.wantAllowed {
				t.Errorf("Authorize() = %v, want %v", allowed, tt.wantAllowed)
			}
		})
	}
}
