package main

import (
	"errors"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"
)

// Example user account
const user address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

// Define a role
const RoleExample boards.Role = "example"

// Define a permission
const PermissionFoo boards.Permission = "foo"

func main() {
	// Define a custom foo permission validation function
	validateFoo := func(_ boards.Permissions, args boards.Args) error {
		// Check that the first argument is the string "bob"
		if name, ok := args[0].(string); !ok || name != "bob" {
			return errors.New("unauthorized")
		}
		return nil
	}

	// Create a permissions instance and assign the custom validator to it
	perms := permissions.New()
	perms.ValidateFunc(PermissionFoo, validateFoo)

	// Add foo permission to example role
	perms.AddRole(RoleExample, PermissionFoo)

	// Add a guest user
	perms.SetUserRoles(user, RoleExample)

	// Call a permissioned callback
	args := boards.Args{"bob"}
	perms.WithPermission(user, PermissionFoo, args, func() {
		println("Hello Bob!")
	})
}

// Output:
// Hello Bob!
