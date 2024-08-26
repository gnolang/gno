// Package accesscontrol implements a role-based access control (RBAC) system for Gno applications.
// It provides functionality to create, assign, revoke, and transfer roles.
//
// Usage:
//
// Import the `gno.land/p/demo/accesscontrol` package to manage roles within your Gno realm. You can create roles,
// assign them to users, revoke them, and transfer role ownership.
//
// Roles can be created by the contract owner using `CreateRole`. Each role is uniquely identified by its name.
//
//	CreateRole("editor")
//
// Use `GrantRole` to assign a role to a user, and `RevokeRole` to remove a role from a user.
//
//	GrantRole("editor", userAddress)
//
//	RevokeRole("editor", userAddress)
//
// Users can renounce their roles using `RenounceRole`, voluntarily removing themselves from a role.
//
//	RenounceRole("editor")
//
// You can look up a role by name with `FindRole`.
//
// FindRole("editor")
//
// Role ownership can be transferred using `SetRoleAdmin`.
//
// SetRoleAdmin(newAdminAddress)
package accesscontrol
