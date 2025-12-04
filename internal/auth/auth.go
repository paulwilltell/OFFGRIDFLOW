// Package auth - auth.go provides role-based authorization.
package auth

import (
"context"
)

// -----------------------------------------------------------------------------
// Authorization Interface
// -----------------------------------------------------------------------------

// Authorizer defines the contract for authorization decisions.
// Implementations can use RBAC, ABAC, or custom authorization logic.
type Authorizer interface {
// Authorize checks if a user can perform an action on a resource.
// Returns true if authorized, false otherwise.
// An error is returned only for authorization system failures.
Authorize(ctx context.Context, user User, action, resource string) (bool, error)
}

// -----------------------------------------------------------------------------
// Standard Actions
// -----------------------------------------------------------------------------

// Standard actions used throughout the system.
const (
ActionRead   = "read"
ActionWrite  = "write"
ActionDelete = "delete"
ActionAdmin  = "admin"
)

// Standard resources used throughout the system.
const (
ResourceEmissions   = "emissions"
ResourceActivities  = "activities"
ResourceCompliance  = "compliance"
ResourceReporting   = "reporting"
ResourceBilling     = "billing"
ResourceUsers       = "users"
ResourceAPIKeys     = "apikeys"
ResourceSettings    = "settings"
)

// -----------------------------------------------------------------------------
// RBAC Authorizer
// -----------------------------------------------------------------------------

// Permission represents a single permission entry.
type Permission struct {
Action   string
Resource string
}

// RBACAuthorizer implements role-based access control.
// It maps roles to sets of permissions and checks if a user's roles
// grant them the requested action on a resource.
type RBACAuthorizer struct {
rolePermissions map[string][]Permission
}

// NewRBACAuthorizer creates an authorizer with default role definitions.
//
// Default roles:
//   - admin: Full access to all resources
//   - editor: Read/write access to core resources, no admin actions
//   - viewer: Read-only access to core resources
func NewRBACAuthorizer() *RBACAuthorizer {
return &RBACAuthorizer{
rolePermissions: map[string][]Permission{
"admin": {
{Action: "*", Resource: "*"}, // Full access
},
"editor": {
{Action: ActionRead, Resource: ResourceEmissions},
{Action: ActionWrite, Resource: ResourceEmissions},
{Action: ActionRead, Resource: ResourceActivities},
{Action: ActionWrite, Resource: ResourceActivities},
{Action: ActionRead, Resource: ResourceCompliance},
{Action: ActionWrite, Resource: ResourceCompliance},
{Action: ActionRead, Resource: ResourceReporting},
{Action: ActionRead, Resource: ResourceSettings},
},
"viewer": {
{Action: ActionRead, Resource: ResourceEmissions},
{Action: ActionRead, Resource: ResourceActivities},
{Action: ActionRead, Resource: ResourceCompliance},
{Action: ActionRead, Resource: ResourceReporting},
},
},
}
}

// NewRBACAuthorizerWithPermissions creates an authorizer with custom role definitions.
func NewRBACAuthorizerWithPermissions(rolePermissions map[string][]Permission) *RBACAuthorizer {
if rolePermissions == nil {
rolePermissions = make(map[string][]Permission)
}
return &RBACAuthorizer{rolePermissions: rolePermissions}
}

// AddRole adds or updates a role with the given permissions.
func (a *RBACAuthorizer) AddRole(role string, permissions []Permission) {
a.rolePermissions[role] = permissions
}

// RemoveRole removes a role from the authorizer.
func (a *RBACAuthorizer) RemoveRole(role string) {
delete(a.rolePermissions, role)
}

// Authorize checks if the user has permission for the action on the resource.
// Returns true if any of the user's roles grant the required permission.
func (a *RBACAuthorizer) Authorize(ctx context.Context, user User, action, resource string) (bool, error) {
// Check primary role
if a.roleHasPermission(user.Role, action, resource) {
return true, nil
}

// Check additional roles
for _, role := range user.Roles {
if a.roleHasPermission(role, action, resource) {
return true, nil
}
}

return false, nil
}

// roleHasPermission checks if a specific role grants the permission.
func (a *RBACAuthorizer) roleHasPermission(role, action, resource string) bool {
permissions, exists := a.rolePermissions[role]
if !exists {
return false
}

for _, perm := range permissions {
if a.matchesPermission(perm, action, resource) {
return true
}
}

return false
}

// matchesPermission checks if a permission grants access for the action/resource.
func (a *RBACAuthorizer) matchesPermission(perm Permission, action, resource string) bool {
// Wildcard permission grants everything
if perm.Action == "*" && perm.Resource == "*" {
return true
}

// Action must match exactly or be wildcard
actionMatch := perm.Action == "*" || perm.Action == action

// Resource must match exactly or be wildcard
resourceMatch := perm.Resource == "*" || perm.Resource == resource

return actionMatch && resourceMatch
}

// GetRolePermissions returns the permissions for a role.
func (a *RBACAuthorizer) GetRolePermissions(role string) []Permission {
perms := a.rolePermissions[role]
if perms == nil {
return nil
}
// Return a copy to prevent external modification
result := make([]Permission, len(perms))
copy(result, perms)
return result
}

// ListRoles returns all defined roles.
func (a *RBACAuthorizer) ListRoles() []string {
roles := make([]string, 0, len(a.rolePermissions))
for role := range a.rolePermissions {
roles = append(roles, role)
}
return roles
}

// -----------------------------------------------------------------------------
// Authorization Helpers
// -----------------------------------------------------------------------------

// CanRead is a convenience function to check read permission.
func CanRead(ctx context.Context, auth Authorizer, user User, resource string) bool {
ok, _ := auth.Authorize(ctx, user, ActionRead, resource)
return ok
}

// CanWrite is a convenience function to check write permission.
func CanWrite(ctx context.Context, auth Authorizer, user User, resource string) bool {
ok, _ := auth.Authorize(ctx, user, ActionWrite, resource)
return ok
}

// CanDelete is a convenience function to check delete permission.
func CanDelete(ctx context.Context, auth Authorizer, user User, resource string) bool {
ok, _ := auth.Authorize(ctx, user, ActionDelete, resource)
return ok
}

// CanAdmin is a convenience function to check admin permission.
func CanAdmin(ctx context.Context, auth Authorizer, user User, resource string) bool {
ok, _ := auth.Authorize(ctx, user, ActionAdmin, resource)
return ok
}
