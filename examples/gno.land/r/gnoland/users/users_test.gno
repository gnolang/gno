package users

import (
	"std"
	"strconv"
	"strings"
	"testing"
)

func TestAddRelease(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/invalid/pkg/path/v1"))
	if strings.Contains(changelog.Latest().Render(), "r/invalid/pkg/path") {
		t.Fatalf("pkgpath was not rejected")
	}
	
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	AddRelease("")

	if !strings.Contains(changelog.Latest().Render(), "r/gnoland/users/v1") {
		t.Fatalf("release was not added")
	}
}
