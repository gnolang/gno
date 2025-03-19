// Package rbac provides a flexible, upgradeable Role-Based Access Control (RBAC)
// system for Gno smart contracts and related applications. It decouples authorization
// logic from fixed addresses, enabling dynamic registration, update, and removal of roles
// and permissions.
//
// ## Overview
//
// The RBAC package encapsulates a manager that maintains an internal registry of roles.
// Each role is defined by a unique name and a set of permissions. A permission is
// represented by a `PermissionChecker` function that validates whether a given caller
// (`std.Address`) satisfies the required access conditions.
//
// Key components of this package include:
//
//  1. **Role**: Represents a role with a name and a collection of permission-checking functions.
//  2. **PermissionChecker**: A function type defined as `func(caller std.Address) error`,
//     used to verify access for a given permission.
//  3. **RBAC Manager**: The core type (RBAC) that encapsulates role registration, permission
//     assignment, verification, updating, and removal.
//
// ## Key Features
//
//   - **Dynamic Role Management**: Roles can be registered, and permissions can be assigned
//     or updated at runtime without requiring contract redeployment.
//   - **Multiple Permissions per Role**: A single role can have multiple permissions,
//     each with its own validation logic.
//   - **Declarative Role Definition**: The package supports a Functional Option pattern,
//     allowing roles and their permissions to be defined declaratively via functions like
//     `DeclareRole` and `WithPermission`.
//   - **Encapsulation**: Internal state (roles registry) is encapsulated within the RBAC
//     manager, preventing unintended external modifications.
//   - **Flexible Validation**: Permission checkers can implement custom logic, supporting
//     arbitrary access control policies.
//
// ## Workflow
//
// Typical usage of the RBAC package includes the following steps:
//
//  1. **Initialization**: Create a new RBAC manager using `NewRBAC()`.
//  2. **Role Registration**: Register roles using `RegisterRole` or declaratively with
//     `DeclareRole`.
//  3. **Permission Assignment**: Add permissions to roles using `RegisterPermission` or the
//     `WithPermission` option during role declaration.
//  4. **Permission Verification**: Validate access by invoking `CheckPermission` with the
//     role name, permission name, and the caller's address (std.Address).
//
// ## Example Usage
//
// The following example demonstrates how to use the RBAC package in both traditional and
// declarative styles:
//
// ```gno
// package main
//
// import (
//
//	"std"
//
//	"gno.land/p/gnoswap/rbac"
//	"gno.land/p/demo/ufmt"
//
// )
//
//	func main() {
//	    // Create a new RBAC manager
//	    manager := rbac.NewRBAC()
//
//	    // Define example addresses
//	    adminAddr := std.Address("admin")
//	    userAddr  := std.Address("user")
//
//	    // --- Traditional Role Registration ---
//	    // Register an "admin" role
//	    if err := manager.RegisterRole("admin"); err != nil {
//	        panic(err)
//	    }
//
//	    // Register an "access" permission for the "admin" role.
//	    // The checker verifies that the caller matches adminAddr.
//	    adminChecker := func(caller std.Address) error {
//	        if caller != adminAddr {
//	            return ufmt.Errorf("caller %s is not admin", caller)
//	        }
//	        return nil
//	    }
//	    if err := manager.RegisterPermission("admin", "access", adminChecker); err != nil {
//	        panic(err)
//	    }
//
//	    // --- Declarative Role Registration ---
//	    // Register an "editor" role with a "modify" permission using the Functional Option pattern.
//	    editorChecker := func(caller std.Address) error {
//	        if caller != userAddr {
//	            return ufmt.Errorf("caller %s is not editor", caller)
//	        }
//	        return nil
//	    }
//	    if err := manager.DeclareRole("editor", rbac.WithPermission("modify", editorChecker)); err != nil {
//	        panic(err)
//	    }
//
//	    // --- Permission Check ---
//	    // Check if adminAddr has the "access" permission on the "admin" role.
//	    if err := manager.CheckPermission("admin", "access", adminAddr); err != nil {
//	        println("Access denied for admin:", err)
//	    } else {
//	        println("Admin access granted")
//	    }
//	}
//
// ```
//
// ## Error Handling
//
// The package reports errors using the ufmt.Errorf function. Typical errors include:
//
// - Registering a role that already exists.
// - Attempting to register a permission for a non-existent role.
// - Verifying a permission that does not exist on a role.
// - Failing a permission check due to a caller not meeting the required conditions.
//
// ## Limitations and Considerations
//
//   - This RBAC implementation does not directly map addresses to roles; instead, it verifies
//     the caller against permission-checking functions registered for a role.
//   - Address validation relies on the logic provided within each PermissionChecker. Ensure that
//     your checkers properly validate `std.Address` values (which follow the Bech32 format).
//   - The encapsulated RBAC manager is designed to minimize external mutation, but integrating it
//     with other modules may require additional mapping between addresses and roles.
//
// # Notes
//
//   - The RBAC system is designed to be upgradeable, enabling contracts to modify permission
//     logic without redeploying the entire contract.
//   - Both imperative and declarative styles are supported, providing flexibility to developers.
//
// Package rbac is intended for use in Gno smart contracts and other systems requiring dynamic,
// upgradeable access control mechanisms.
package rbac
