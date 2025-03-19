package rbac

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	prbac "gno.land/p/gnoswap/rbac"
)

var (
	testCaller  = testutils.TestAddress("caller")
	wrongCaller = testutils.TestAddress("wrong")
)

func TestRegisterRole(t *testing.T) {
	err := RegisterRole("admin")
	uassert.NoError(t, err)

	// Fail: already existed role
	err = RegisterRole("admin")
	uassert.Error(t, err)
}

func TestRegisterAndCheckPermission(t *testing.T) {
	roleName := "moderator"
	permissionName := "can_delete"

	// register role
	err := RegisterRole(roleName)
	uassert.NoError(t, err)

	// register permission
	checker := func(addr std.Address) error {
		if addr == testCaller {
			return nil
		}
		return errors.New("wrong caller")
	}
	err = RegisterPermission(roleName, permissionName, checker)
	uassert.NoError(t, err)

	// check permission - success case
	err = CheckPermission(roleName, permissionName, testCaller)
	uassert.NoError(t, err)

	// check permission - fail case (wrong caller)
	err = CheckPermission(roleName, permissionName, wrongCaller)
	uassert.Error(t, err)
}

func TestUpdateAndRemovePermission(t *testing.T) {
	roleName := "editor"
	permissionName := "can_edit"

	// initial setup
	RegisterRole(roleName)
	originalChecker := func(addr std.Address) error {
		if addr == testCaller {
			return nil
		}
		return errors.New("wrong caller")
	}
	RegisterPermission(roleName, permissionName, originalChecker)

	// update permission
	newChecker := func(addr std.Address) error {
		return errors.New("wrong caller")
	}
	err := UpdatePermission(roleName, permissionName, newChecker)
	uassert.NoError(t, err)

	// check updated permission
	err = CheckPermission(roleName, permissionName, testCaller)
	uassert.Error(t, err)

	// remove permission
	err = RemovePermission(roleName, permissionName)
	uassert.NoError(t, err)

	// check removed permission
	err = CheckPermission(roleName, permissionName, testCaller)
	uassert.Error(t, err)
}

func TestDeclareRole(t *testing.T) {
	roleName := "supervisor"
	permissionName := "can_supervise"

	checker := func(addr std.Address) error {
		if addr == testCaller {
			return nil
		}
		return errors.New("wrong caller")
	}

	err := DeclareRole(roleName,
		prbac.WithPermission(permissionName, checker))
	uassert.NoError(t, err)

	// check permission
	err = CheckPermission(roleName, permissionName, testCaller)
	uassert.NoError(t, err)
}
