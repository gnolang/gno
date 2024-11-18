package upgradable_testing

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"

	"gno.land/r/demo/upgradable/admin"
	"gno.land/r/demo/upgradable/logic/v1"
	"gno.land/r/demo/upgradable/logic/v2"
	"gno.land/r/demo/upgradable/store"

)

func TestPackage(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // XXX: should not need this

	store.Init()

	v1.Init()
	urequire.Equal(t, admin.ReadCounter(), uint64(0))
	admin.UpdateCounter()
	urequire.Equal(t, admin.ReadCounter(), uint64(1))

	v2.Init()
	urequire.Equal(t, admin.ReadCounter(), uint64(1))
	admin.UpdateCounter()
	urequire.Equal(t, admin.ReadCounter(), uint64(3))

	urequire.Equal(t, admin.DebugPrevRealm(), admin.DebugAdminPrevRealm())
}