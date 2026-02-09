# Boards2 Permissions

This realm provides a custom `gno.land/p/gnoland/boards` package `Permissions` implementation.

This implementation is used by default when creating non open boards, to organize board members,
roles and permissions. It uses an underlying DAO to manage users and roles.

The `Permissions` type also supports optionally setting validation functions to be triggered within
`WithPermission()` method before a permissioned callback is called.

Permissioned call example:

```go
import (
  "errors"

  "gno.land/p/gnoland/boards"

  "gno.land/r/gnoland/boards/permissions"
)

// Define a foo permissions
const PermissionFoo boards.Permissions = "foo"

// Define a custom foo permission validation function
validateFoo := func(_ boards.Permissions, args boards.Args) error {
    if name, ok := args[0].(string); !ok || name != "bar" {
        return errors.New("unauthorized")
    }
    return nil
}

// Create a permissions instance and assign the custom validator to it
perms := permissions.New()
perms.ValidateFunc(PermisionFoo, validateFoo)

// Call a permissioned callback
args := boards.Args{"bar"}
user := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
perms.WithPermission(cross, user, PermisionFoo, args, func(realm) {
    println("Hello bar!")
})
```
