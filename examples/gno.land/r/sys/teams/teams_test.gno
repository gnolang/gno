package teams

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestPackage(t *testing.T) {
	if GetAdmins("gnoland") != nil {
		t.Errorf("gnoland should have no admins")
	}
	if GetEditors("gnoland") != nil {
		t.Errorf("gnoland should have no admins")
	}

	std.TestSetPrevRealm("r/gnoland/home")
	// std.TestSetCurrentRealm("r/gnoland/home")
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test2)
	SetMembers(
		string(test1), string(RoleAdmin),
	)
}
