# Boards Permissions Extension

This is a `gno.land/p/gnoland/boards` package extension that provides a custom
`Permissions` implementation that uses an underlying DAO to manage users and
roles.

It also supports optionally setting validation functions to be triggered by the
`WithPermission()` method before a callback is called. Validators allows adding
custom checks and requirements before the callback is called.

Usage Example:

```go
import (
  "errors"

  "gno.land/p/gnoland/boards"
  "gno.land/p/gnoland/boards/exts/permissions"
)

// Example user account
const user address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

// Define a foo permissions
const PermissionFoo boards.Permissions = "foo"

// Define a custom foo permission validation function
validateFoo := func(_ boards.Permissions, args boards.Args) error {
    // Check that the first argument is the string "bar"
    if name, ok := args[0].(string); !ok || name != "bar" {
        return errors.New("unauthorized")
    }
    return nil
}

// Create a permissions instance and assign the custom validator to it
perms := permissions.New()
perms.ValidateFunc(PermisionFoo, validateFoo)

// Add foo permission to guests
perms.AddRole(permissions.RoleGuest, PermissionFoo)

// Add a guest user
perms.SetUserRoles(cross, user, permissions.RoleGuest)

// Call a permissioned callback
args := boards.Args{"bar"}
user := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
perms.WithPermission(user, PermisionFoo, args, func() {
    println("Hello bar!")
})
```
