package teams

import (
	"std"
	"testing"

	"gno.land/r/sys/teams"
)

func TestPackage(t *testing.T) {
	if teams.GetAdmins("gnoland") != nil {
		t.Errorf("gnoland should have no admins")
	}
	if teams.GetEditors("gnoland") != nil {
		t.Errorf("gnoland should have no admins")
	}
	std.TestSetPrevRealm("r/gnoland/home")
	teams.SetMembers(
	// "g123456789", teams.RoleAdmin,
	)
}
