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
		t.Errorf("gnoland should have no editors")
	}

	std.TestSetRealm(std.NewCodeRealm("gno.land/r/gnoland/home"))
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")
	SetMembers(
		string(test1), string(RoleAdmin),
		string(test2), string(RoleEditor),
	)

	admins := GetAdmins("gnoland")
	if !isEqualAddressSlice(admins, []std.Address{test1}) {
		t.Errorf("gnoland should have test1 as admins")
	}
	editors := GetEditors("gnoland")
	if !isEqualAddressSlice(editors, []std.Address{test1, test2}) {
		t.Errorf("gnoland should have test1 and test2 as editors")
	}
}

func isEqualAddressSlice(a, b []std.Address) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
