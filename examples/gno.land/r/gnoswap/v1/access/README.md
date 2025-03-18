# Access Control

The `access` package provides a configuration-based wrapper around the `p/rbac` package, offering simplified role management and access control for Gno smart contracts.

## Key Features

- **Configuration-based Setup**: Initialize access control with a simple configuration containing role-to-address mappings
- **Predefined Roles**: Built-in roles for common access patterns (admin, governance, router, pool, etc.)
- **Dynamic Role Management**: Support for creating new roles and updating role addresses at runtime
- **Simple Permission Checks**: Utility functions for checking role-based permissions

## Predefined Roles

| Role Name | Value | Description |
|-----------|-------|-------------|
| `ROLE_ADMIN` | `admin` | Admin role |
| `ROLE_GOVERNANCE` | `governance` | Governance role |
| `ROLE_GOV_STAKER` | `gov_staker` | Governance staker role |
| `ROLE_ROUTER` | `router` | Router role |
| `ROLE_POOL` | `pool` | Pool role |
| `ROLE_POSITION` | `position` | Position role |
| `ROLE_STAKER` | `staker` | Staker role |
| `ROLE_LAUNCHPAD` | `launchpad` | Launchpad role |
| `ROLE_EMISSION` | `emission` | Emission role |

## API Overview

### Configuration

```go
type Config struct {
    Roles map[string]std.Address
}

// Set configuration
func SetConfig(cfg *Config) error

// Get current configuration
func GetCurrentConfig() *Config
```

### Role Management

```go
// Set or update a role with address
func SetRole(roleName string, address std.Address) error

// Create a new role with address
func CreateRole(roleName string, address std.Address) error

// Update address for a specific role
func UpdateRoleAddress(roleName string, newAddress std.Address) error

// Check if a role exists
func RoleExists(roleName string) bool

// Get all registered roles
func GetRoles() []string
```

### Permission Checks

`XXXOnly` functions are used to check if the caller has the given role.

It follows the pattern of `XXXOnly(caller std.Address, newXXX ...std.Address) error`.

For example, `AdminOnly` function is defined as follows:

```go
// Check if caller has admin role
func AdminOnly(caller std.Address, newAdmin ...std.Address) error
```

Parameters:

- `caller`: The address to check for role permission
- `newAddress` (optional): New address to update the role with

Example usage:

```go
// Simple permission check
if err := access.AdminOnly(callerAddr); err != nil {
    return err
}

// Check permission and update role address
if err := access.AdminOnly(currentAdmin, newAdminAddr); err != nil {
    return err
}
```

## Implementation Details

1. **Configuration**: The package maintains a global configuration storing role-to-address mappings.

2. **Permission Checking**: Each role is associated with an address checker that validates if a caller matches the configured address.

3. **Role Management**:
   - Roles can be pre-configured during initialization
   - New roles can be created at runtime
   - Role addresses can be updated dynamically

The sequence diagram above illustrates the flow of initialization and role management operations in the Access Control system.

## Limitations

- Only supports single-address to role mapping
- All roles use the same permission type ("access")
- Configuration must be initialized before using any functionality
- Global configuration state may need careful management in complex applications
