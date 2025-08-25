// Package accesscontrol implements a role-based access control (RBAC) system for Gno applications.
// It provides functionality to create, assign, revoke, and transfer roles.
//
// # Usage
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
//
// Key events
// - `RoleCreatedEvent`: Triggered when a new role is created
// Key includes:
// `roleName` (name of the role)
// `sender` (address of the sender)
//
// - `RoleGrantedEvent`: Triggered when a role is granted to an account
// Key includes:
// `roleName` (name of the role)
// `account` (address of the account)
// `sender` (address of the sender)
//
// - `RoleRevokedEvent`: Triggered when a role is revoked from an account
// Key includes:
// `roleName` (name of the role)
// `account` (address of the account)
// `sender` (address of the sender)
//
// - `RoleRenouncedEvent`: Triggered when a role is renounced by an account
// Key includes:
// `roleName` (name of the role)
// `account` (address of the account)
//
// - `RoleSetEvent`: Triggered when a role's administrator is set or changed
// Key includes:
// `roleName` (name of the role)
// `newAdmin` (address of the new administrator)
// `sender` (address of the sender)
package accesscontrol
