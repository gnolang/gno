package rbac

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	prbac "gno.land/p/gnoswap/rbac"
)

var (
	testCaller      = testutils.TestAddress("caller")
	wrongCaller     = testutils.TestAddress("wrong")
	testCallerRealm = std.NewUserRealm(testCaller)
)

func TestRegisterRole(t *testing.T) {
	tt := []struct {
		name          string
		caller        std.Address
		owner         std.Address
		defaultRole   string
		inputRole     string
		expectedError error
	}{
		{
			name:          "register role success",
			caller:        testCaller,
			owner:         testCaller,
			defaultRole:   "test_role",
			inputRole:     "pool",
			expectedError: nil,
		},
		{
			name:          "register role failed by wrong caller",
			caller:        wrongCaller,
			owner:         testCaller,
			defaultRole:   "test_role",
			inputRole:     "pool",
			expectedError: prbac.ErrUnauthorized,
		},
		{
			name:          "register role failed by already registered role",
			caller:        testCaller,
			owner:         testCaller,
			defaultRole:   "test_role",
			inputRole:     "test_role",
			expectedError: ufmt.Errorf("role test_role already exists"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// given
			testing.SetOriginCaller(tc.caller)
			globalManager = prbac.NewRBACWithAddress(tc.owner)

			// when
			err := RegisterRole(cross, tc.inputRole)

			// then
			uassert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestRegisterPermission(t *testing.T) {
	checkerFn := func(addr std.Address) error {
		if addr == testCaller {
			return nil
		}
		return errors.New("checker failed")
	}

	tt := []struct {
		name                string
		caller              std.Address
		owner               std.Address
		inputRoleName       string
		inputPermissionName string
		inputChecker        func(addr std.Address) error
		expectedError       error
	}{
		{
			name:                "register permission success",
			caller:              testCaller,
			owner:               testCaller,
			inputRoleName:       "test_role",
			inputPermissionName: "can_delete",
			inputChecker:        checkerFn,
			expectedError:       nil,
		},
		{
			name:                "register permission failed by wrong caller",
			caller:              wrongCaller,
			owner:               testCaller,
			inputRoleName:       "test_role",
			inputPermissionName: "can_delete",
			inputChecker:        checkerFn,
			expectedError:       prbac.ErrUnauthorized,
		},
		{
			name:                "register permission failed by not existing role",
			caller:              testCaller,
			owner:               testCaller,
			inputRoleName:       "wrong_role",
			inputPermissionName: "can_delete",
			inputChecker:        checkerFn,
			expectedError:       ufmt.Errorf("role wrong_role does not exist"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// given
			testing.SetOriginCaller(tc.owner)
			globalManager = prbac.NewRBACWithAddress(tc.owner)

			// given - input role
			err := RegisterRole(cross, tc.inputRoleName)
			uassert.NoError(t, err)

			// when
			testing.SetOriginCaller(tc.caller)
			err = RegisterPermission(cross, tc.inputRoleName, tc.inputPermissionName, tc.inputChecker)

			// then
			uassert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestCheckPermission(t *testing.T) {
	checkerFn := func(addr std.Address) error {
		if addr == testCaller {
			return nil
		}
		return errors.New("checker failed")
	}

	tt := []struct {
		name                string
		caller              std.Address
		owner               std.Address
		roleName            string
		permissionName      string
		checker             func(addr std.Address) error
		inputRoleName       string
		inputPermissionName string
		expectedError       error
	}{
		{
			name:                "check permission success",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			checker:             checkerFn,
			inputRoleName:       "test_role",
			inputPermissionName: "can_delete",
			expectedError:       nil,
		}, {
			name:                "check permission failed by wrong role",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			checker:             checkerFn,
			inputRoleName:       "wrong_role",
			inputPermissionName: "can_delete",
			expectedError:       errors.New("role wrong_role does not exist"),
		}, {
			name:                "check permission failed by wrong permission",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			checker:             checkerFn,
			inputRoleName:       "test_role",
			inputPermissionName: "wrong_permission",
			expectedError:       errors.New("permission wrong_permission does not exist for role test_role"),
		}, {
			name:                "check permission failed by has no permission",
			caller:              wrongCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			checker:             checkerFn,
			inputRoleName:       "test_role",
			inputPermissionName: "can_delete",
			expectedError:       errors.New("checker failed"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// given
			testing.SetOriginCaller(tc.owner)
			globalManager = prbac.NewRBACWithAddress(tc.owner)

			// given - input role
			err := RegisterRole(cross, tc.roleName)
			uassert.NoError(t, err)

			// given - input permission
			err = RegisterPermission(cross, tc.roleName, tc.permissionName, tc.checker)
			uassert.NoError(t, err)

			// when
			err = CheckPermission(tc.inputRoleName, tc.inputPermissionName, tc.caller)

			// then
			uassert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestUpdatePermission(t *testing.T) {
	checkerFn := func(addr std.Address) error {
		if addr == testCaller {
			return nil
		}
		return errors.New("checker failed")
	}

	tt := []struct {
		name                string
		caller              std.Address
		owner               std.Address
		roleName            string
		permissionName      string
		checker             func(addr std.Address) error
		inputRoleName       string
		inputPermissionName string
		inputChecker        func(addr std.Address) error
		expectedError       error
		expectedCheckError  error
	}{
		{
			name:                "update permission success",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			checker:             checkerFn,
			inputRoleName:       "test_role",
			inputPermissionName: "can_delete",
			inputChecker: func(addr std.Address) error {
				if addr == testCaller {
					return nil
				}
				return errors.New("new checker failed")
			},
			expectedError:      nil,
			expectedCheckError: errors.New("new checker failed"),
		},
		{
			name:                "update permission failed by not existing role",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			checker:             checkerFn,
			inputRoleName:       "wrong_role",
			inputPermissionName: "can_delete",
			inputChecker: func(addr std.Address) error {
				if addr == testCaller {
					return nil
				}
				return errors.New("new checker failed")
			},
			expectedError:      ufmt.Errorf("role wrong_role does not exist"),
			expectedCheckError: nil,
		},
		{
			name:                "update permission failed by not existing permission",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			checker:             checkerFn,
			inputRoleName:       "test_role",
			inputPermissionName: "wrong_permission",
			inputChecker: func(addr std.Address) error {
				if addr == testCaller {
					return nil
				}
				return errors.New("new checker failed")
			},
			expectedError:      ufmt.Errorf("permission wrong_permission does not exist for role test_role"),
			expectedCheckError: nil,
		},
		{
			name:                "update permission failed by wrong caller",
			caller:              wrongCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			checker:             checkerFn,
			inputRoleName:       "test_role",
			inputPermissionName: "can_delete",
			inputChecker: func(addr std.Address) error {
				if addr == testCaller {
					return nil
				}
				return errors.New("new checker failed")
			},
			expectedError:      prbac.ErrUnauthorized,
			expectedCheckError: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// given
			testing.SetOriginCaller(tc.owner)
			globalManager = prbac.NewRBACWithAddress(tc.owner)

			// given - input role
			err := RegisterRole(cross, tc.roleName)
			uassert.NoError(t, err)

			// given - input permission
			err = RegisterPermission(cross, tc.roleName, tc.permissionName, tc.checker)
			uassert.NoError(t, err)

			// when
			testing.SetOriginCaller(tc.caller)
			err = UpdatePermission(cross, tc.inputRoleName, tc.inputPermissionName, tc.inputChecker)

			// then
			uassert.ErrorIs(t, err, tc.expectedError)

			// then
			err = CheckPermission(tc.inputRoleName, tc.inputPermissionName, tc.caller)
			uassert.ErrorIs(t, err, tc.expectedCheckError)
		})
	}
}

func TestRemovePermission(t *testing.T) {
	tt := []struct {
		name                string
		caller              std.Address
		owner               std.Address
		roleName            string
		permissionName      string
		inputRoleName       string
		inputPermissionName string
		expectedError       error
	}{
		{
			name:                "remove permission success",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			inputRoleName:       "test_role",
			inputPermissionName: "can_delete",
			expectedError:       nil,
		},
		{
			name:                "remove permission failed by wrong caller",
			caller:              wrongCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			inputRoleName:       "test_role",
			inputPermissionName: "can_delete",
			expectedError:       prbac.ErrUnauthorized,
		},
		{
			name:                "remove permission failed by not existing role",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			inputRoleName:       "wrong_role",
			inputPermissionName: "can_delete",
			expectedError:       ufmt.Errorf("cannot remove permission from non-existent role wrong_role"),
		},
		{
			name:                "remove permission failed by not existing permission",
			caller:              testCaller,
			owner:               testCaller,
			roleName:            "test_role",
			permissionName:      "can_delete",
			inputRoleName:       "test_role",
			inputPermissionName: "wrong_permission",
			expectedError:       ufmt.Errorf("permission wrong_permission does not exist for role test_role"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// given
			testing.SetOriginCaller(tc.owner)
			globalManager = prbac.NewRBACWithAddress(tc.owner)

			// given - input role
			err := RegisterRole(cross, tc.roleName)
			uassert.NoError(t, err)

			// given - input permission
			checkerFn := func(addr std.Address) error {
				if addr == testCaller {
					return nil
				}
				return errors.New("checker failed")
			}
			err = RegisterPermission(cross, tc.roleName, tc.permissionName, checkerFn)
			uassert.NoError(t, err)

			// when
			testing.SetOriginCaller(tc.caller)
			err = RemovePermission(cross, tc.inputRoleName, tc.inputPermissionName)

			// then
			uassert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestRemoveRole(t *testing.T) {
	tt := []struct {
		name          string
		caller        std.Address
		owner         std.Address
		roleName      string
		inputRoleName string
		expectedError error
	}{
		{
			name:          "remove role success",
			caller:        testCaller,
			owner:         testCaller,
			roleName:      "test_role",
			inputRoleName: "test_role",
			expectedError: nil,
		},
		{
			name:          "remove role failed by wrong caller",
			caller:        wrongCaller,
			owner:         testCaller,
			roleName:      "test_role",
			inputRoleName: "test_role",
			expectedError: prbac.ErrUnauthorized,
		},
		{
			name:          "remove role failed by not existing role",
			caller:        testCaller,
			owner:         testCaller,
			roleName:      "test_role",
			inputRoleName: "wrong_role",
			expectedError: ufmt.Errorf("role wrong_role does not exist"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// given
			testing.SetOriginCaller(tc.owner)
			globalManager = prbac.NewRBACWithAddress(tc.owner)

			// given - input role
			err := RegisterRole(cross, tc.roleName)
			uassert.NoError(t, err)

			// when
			testing.SetOriginCaller(tc.caller)
			err = RemoveRole(cross, tc.inputRoleName)

			// then
			uassert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestDeclareRole(t *testing.T) {
	tt := []struct {
		name             string
		caller           std.Address
		owner            std.Address
		roleName         string
		inputRoleName    string
		inputRoleOptions []prbac.RoleOption
		expectedError    error
	}{
		{
			name:          "declare role success",
			caller:        testCaller,
			owner:         testCaller,
			roleName:      "test_role",
			inputRoleName: "newAdmin",
			inputRoleOptions: []prbac.RoleOption{
				prbac.WithPermission("can_delete", func(addr std.Address) error {
					return nil
				}),
			},
			expectedError: nil,
		},
		{
			name:          "declare role failed by wrong caller",
			caller:        wrongCaller,
			owner:         testCaller,
			roleName:      "test_role",
			inputRoleName: "newAdmin",
			inputRoleOptions: []prbac.RoleOption{
				prbac.WithPermission("can_delete", func(addr std.Address) error {
					return nil
				}),
			},
			expectedError: prbac.ErrUnauthorized,
		},
		{
			name:          "declare role failed by already declared role",
			caller:        wrongCaller,
			owner:         testCaller,
			roleName:      "test_role",
			inputRoleName: "test_role",
			inputRoleOptions: []prbac.RoleOption{
				prbac.WithPermission("can_delete", func(addr std.Address) error {
					return nil
				}),
			},
			expectedError: ufmt.Errorf("role test_role already exists"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// given
			testing.SetOriginCaller(tc.owner)
			globalManager = prbac.NewRBACWithAddress(tc.owner)

			// when
			testing.SetOriginCaller(tc.caller)
			err := DeclareRole(cross, tc.inputRoleName, tc.inputRoleOptions...)

			// then
			uassert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
