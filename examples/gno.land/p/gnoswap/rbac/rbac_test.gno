package rbac

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

var (
	adminRole  = "admin"
	editorRole = "editor"

	accessPermission = "access"
	editPermission   = "edit"
	readPermission   = "read"
	writePermission  = "write"
	deletePermission = "delete"

	adminAddr  = testutils.TestAddress(adminRole)
	userAddr   = testutils.TestAddress("user")
	editorAddr = testutils.TestAddress(editorRole)
)

func adminChecker(caller std.Address) error {
	if caller != adminAddr {
		return ufmt.Errorf("caller is not admin")
	}
	return nil
}

func editorChecker(caller std.Address) error {
	if caller != editorAddr {
		return ufmt.Errorf("caller is not editor")
	}
	return nil
}

func TestRegisterRole(t *testing.T) {
	manager := New()

	err := manager.RegisterRole(adminRole)
	uassert.NoError(t, err)

	err = manager.RegisterRole(adminRole)
	uassert.Error(t, err)
}

func TestRegisterPermission(t *testing.T) {
	manager := New()

	err := manager.RegisterPermission("nonexistent", accessPermission, adminChecker)
	uassert.Error(t, err)

	// check permission
	err = manager.RegisterRole(adminRole)
	uassert.NoError(t, err)

	err = manager.RegisterPermission(adminRole, accessPermission, adminChecker)
	uassert.NoError(t, err)
}

func TestCheckPermission(t *testing.T) {
	manager := New()

	err := manager.RegisterRole(adminRole)
	uassert.NoError(t, err)

	err = manager.RegisterPermission(adminRole, accessPermission, adminChecker)
	uassert.NoError(t, err)

	// success: calling from admin address
	err = manager.CheckPermission(adminRole, accessPermission, adminAddr)
	uassert.NoError(t, err)

	// failure: calling from user address
	err = manager.CheckPermission(adminRole, accessPermission, userAddr)
	uassert.Error(t, err)
}

func TestUpdatePermission(t *testing.T) {
	manager := New()

	err := manager.RegisterRole(adminRole)
	uassert.NoError(t, err)

	err = manager.RegisterPermission(adminRole, accessPermission, adminChecker)
	uassert.NoError(t, err)

	newChecker := func(caller std.Address) error {
		if caller != userAddr {
			return ufmt.Errorf("caller is not user")
		}
		return nil
	}

	err = manager.UpdatePermission(adminRole, accessPermission, newChecker)
	uassert.NoError(t, err)

	err = manager.CheckPermission(adminRole, accessPermission, userAddr)
	uassert.NoError(t, err)

	err = manager.CheckPermission(adminRole, accessPermission, adminAddr)
	uassert.Error(t, err)
}

func TestRemovePermission(t *testing.T) {
	manager := New()

	err := manager.RegisterRole(adminRole)
	uassert.NoError(t, err)

	err = manager.RegisterPermission(adminRole, accessPermission, adminChecker)
	uassert.NoError(t, err)

	// success: remove permission
	err = manager.RemovePermission(adminRole, accessPermission)
	uassert.NoError(t, err)

	err = manager.CheckPermission(adminRole, accessPermission, adminAddr)
	uassert.Error(t, err)

	// double remove
	err = manager.RemovePermission(adminRole, accessPermission)
	uassert.Error(t, err)
}

func TestDeclareRole(t *testing.T) {
	manager := New()

	err := manager.DeclareRole(editorRole, WithPermission(editPermission, editorChecker))
	uassert.NoError(t, err)

	// only editor can edit
	err = manager.CheckPermission(editorRole, editPermission, editorAddr)
	uassert.NoError(t, err)

	// no permission for admin and user
	err = manager.CheckPermission(editorRole, editPermission, adminAddr)
	uassert.Error(t, err)

	err = manager.CheckPermission(editorRole, editPermission, userAddr)
	uassert.Error(t, err)
}

func TestApplyMultiplePermissions(t *testing.T) {
	manager := New()

	err := manager.RegisterRole(adminRole)
	uassert.NoError(t, err)

	permissions := map[string]func(std.Address) error{
		readPermission: func(caller std.Address) error {
			if caller != adminAddr {
				return ufmt.Errorf("only admin can read")
			}
			return nil
		},
		writePermission: func(caller std.Address) error {
			if caller != adminAddr {
				return ufmt.Errorf("only admin can write")
			}
			return nil
		},
		deletePermission: func(caller std.Address) error {
			if caller != adminAddr {
				return ufmt.Errorf("only admin can delete")
			}
			return nil
		},
	}

	// Register all permissions
	for perm, checker := range permissions {
		err := manager.RegisterPermission(adminRole, perm, checker)
		uassert.NoError(t, err)
	}

	tests := []struct {
		name        string
		role        string
		permission  string
		caller      std.Address
		expectError bool
	}{
		{"admin can read", adminRole, readPermission, adminAddr, false},
		{"admin can write", adminRole, writePermission, adminAddr, false},
		{"admin can delete", adminRole, deletePermission, adminAddr, false},
		{"admin cannot use editor permission", editorRole, readPermission, adminAddr, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CheckPermission(tt.role, tt.permission, tt.caller)
			if tt.expectError {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestRemoveRole(t *testing.T) {
	manager := New()

	t.Run("success remove non-system role", func(t *testing.T) {
		// First register role and permission
		err := manager.RegisterRole(editorRole)
		uassert.NoError(t, err)

		err = manager.RegisterPermission(editorRole, editPermission, editorChecker)
		uassert.NoError(t, err)

		// Check if permission is set correctly
		err = manager.CheckPermission(editorRole, editPermission, editorAddr)
		uassert.NoError(t, err)

		// Remove role
		err = manager.RemoveRole(editorRole)
		uassert.NoError(t, err)

		// Check if role is removed
		err = manager.CheckPermission(editorRole, editPermission, editorAddr)
		uassert.Error(t, err)
	})

	t.Run("fail to remove non-existent role", func(t *testing.T) {
		err := manager.RemoveRole("non_existent_role")
		uassert.Error(t, err)
	})

	t.Run("fail to remove system role", func(t *testing.T) {
		// Register system role
		err := manager.RegisterRole(adminRole)
		uassert.NoError(t, err)

		err = manager.RegisterPermission(adminRole, accessPermission, adminChecker)
		uassert.NoError(t, err)

		// Try to remove system role
		err = manager.RemoveRole(adminRole)
		uassert.Error(t, err)

		// Check if system role still exists
		err = manager.CheckPermission(adminRole, accessPermission, adminAddr)
		uassert.NoError(t, err)
	})

	t.Run("remove role with multiple permissions", func(t *testing.T) {
		// Create role with multiple permissions
		multiRole := "multi_role"
		err := manager.RegisterRole(multiRole)
		uassert.NoError(t, err)

		// Register multiple permissions
		permissions := map[string]func(std.Address) error{
			readPermission: func(caller std.Address) error {
				if caller != userAddr {
					return ufmt.Errorf("only user can read")
				}
				return nil
			},
			writePermission: func(caller std.Address) error {
				if caller != userAddr {
					return ufmt.Errorf("only user can write")
				}
				return nil
			},
		}

		for perm, checker := range permissions {
			err = manager.RegisterPermission(multiRole, perm, checker)
			uassert.NoError(t, err)
		}

		// Check if all permissions are set correctly
		for perm := range permissions {
			err = manager.CheckPermission(multiRole, perm, userAddr)
			uassert.NoError(t, err)
		}

		// Remove role
		err = manager.RemoveRole(multiRole)
		uassert.NoError(t, err)

		// Check if all permissions are removed
		for perm := range permissions {
			err = manager.CheckPermission(multiRole, perm, userAddr)
			uassert.Error(t, err)
		}
	})
}
