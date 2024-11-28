package upgradable_testing

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"

	"gno.land/r/x/manfred_upgrade_patterns/upgrade_g/admin"
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_g/logic/v1"
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_g/logic/v2"
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_g/store"

)

func TestPackage(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // XXX: should not need this

	store.Init()

	v1.Init()
	urequire.Equal(t, uint64(0), admin.ReadCounter())
	admin.UpdateCounter()
	urequire.Equal(t, uint64(1), admin.ReadCounter())
                   
	v2.Init()         
	urequire.Equal(t, uint64(1), admin.ReadCounter())
	admin.UpdateCounter()
	urequire.Equal(t, uint64(3), admin.ReadCounter())

	urequire.Equal(t, admin.DebugPrevRealm(), admin.DebugAdminPrevRealm())
}