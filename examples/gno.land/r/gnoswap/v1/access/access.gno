package access

import (
	"std"

	"gno.land/p/demo/ufmt"

	prbac "gno.land/p/gnoswap/rbac"
	"gno.land/r/gnoswap/v1/rbac"
)

// event types
const (
	roleSet     = "RoleSet"
	roleUpdated = "RoleUpdated"
	roleCreated = "RoleCreated"
)

var currentConfig *config

// set up the default roles and current config
func init() {
	cfg, err := initializeDefaultRoles()
	if err != nil {
		panic(err)
	}
	currentConfig = cfg
}

// config contains all the addresses needed for access control
type config struct {
	roles map[string]std.Address
}

// newConfig returns an empty config with zero addresses
func newConfig() *config {
	return &config{
		roles: make(map[string]std.Address),
	}
}

// validateConfig checks if the configuration state is valid
func validateConfig() error {
	if currentConfig == nil {
		return ufmt.Errorf(errNotInitialized)
	}
	return nil
}

// validateRoleName checks if the role name is valid
func validateRoleName(roleName string) error {
	if roleName == "" {
		return ufmt.Errorf(errEmptyRole)
	}
	return nil
}

// createAddressChecker creates a permission checker for a specific address
func createAddressChecker(addr std.Address, roleName string) prbac.PermissionChecker {
	return func(caller std.Address) error {
		if caller != addr {
			return ufmt.Errorf(errUnauthorized, caller.String(), roleName)
		}
		return nil
	}
}

// SetConfig sets the current configuration
func SetConfig(cfg *config) error {
	if cfg == nil {
		return ufmt.Errorf(errConfigNil)
	}
	currentConfig = cfg
	return nil
}

// initialize registers all roles and permissions to the global RBAC manager
func initialize(cfg *config) error {
	if cfg == nil {
		return ufmt.Errorf(errConfigNil)
	}

	for roleName, addr := range cfg.roles {
		if err := declareNewRole(roleName, addr); err != nil {
			return err
		}
	}

	return SetConfig(cfg)
}

// declareNewRole declares a new role with permissions
func declareNewRole(roleName string, addr std.Address) error {
	checker := createAddressChecker(addr, roleName)
	if err := rbac.DeclareRole(roleName, prbac.WithPermission(PERM_ACCESS, checker)); err != nil {
		return ufmt.Errorf(errDeclareRole, roleName, err)
	}
	return nil
}

// GetCurrentConfig returns the current configuration
func GetCurrentConfig() *config {
	if currentConfig == nil {
		return nil
	}
	return currentConfig
}

// SetRole sets or updates a role with the given address
func SetRole(roleName string, address std.Address) error {
	if err := validateConfig(); err != nil {
		return err
	}

	if err := validateRoleName(roleName); err != nil {
		return err
	}

	prevAddress := currentConfig.roles[roleName]
	checker := createAddressChecker(address, roleName)

	if RoleExists(roleName) {
		if err := rbac.UpdatePermission(roleName, PERM_ACCESS, checker); err != nil {
			return err
		}
	} else {
		if err := rbac.DeclareRole(roleName, prbac.WithPermission(PERM_ACCESS, checker)); err != nil {
			return err
		}
	}

	currentConfig.roles[roleName] = address

	std.Emit(
		roleSet,
		"roleName", roleName,
		"prevAddress", prevAddress.String(),
		"newAddress", address.String(),
	)
	return nil
}

func GetAddress(roleName string) (std.Address, error) {
	if currentConfig == nil {
		return std.Address(""), ufmt.Errorf(errNotInitialized)
	}
	return currentConfig.roles[roleName], nil
}

// UpdateRoleAddress updates the address for a specific role's permission checker
func UpdateRoleAddress(roleName string, newAddress std.Address) error {
	if err := validateConfig(); err != nil {
		return err
	}

	prevAddress := currentConfig.roles[roleName]
	currentConfig.roles[roleName] = newAddress
	newChecker := createAddressChecker(newAddress, roleName)

	std.Emit(
		roleUpdated,
		"roleName", roleName,
		"prevAddress", prevAddress.String(),
		"newAddress", newAddress.String(),
	)
	return rbac.UpdatePermission(roleName, PERM_ACCESS, newChecker)
}

// CreateRole creates a new role with the given address
func CreateRole(roleName string, address std.Address) error {
	if err := validateConfig(); err != nil {
		return err
	}

	if err := validateRoleName(roleName); err != nil {
		return err
	}

	if RoleExists(roleName) {
		return ufmt.Errorf(errRoleExists, roleName)
	}

	if err := declareNewRole(roleName, address); err != nil {
		return err
	}

	currentConfig.roles[roleName] = address

	std.Emit(
		roleCreated,
		"roleName", roleName,
		"address", address.String(),
	)
	return nil
}

// RoleExists checks if a role exists in the current configuration
func RoleExists(roleName string) bool {
	if currentConfig == nil {
		return false
	}
	_, exists := currentConfig.roles[roleName]
	return exists
}

// GetRoles returns all registered role names
func GetRoles() []string {
	if currentConfig == nil {
		return []string{}
	}
	roles := make([]string, 0, len(currentConfig.roles))
	for role := range currentConfig.roles {
		roles = append(roles, role)
	}
	return roles
}
