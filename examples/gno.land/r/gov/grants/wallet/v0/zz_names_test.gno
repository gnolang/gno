package wallet

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"

	"gno.land/r/sys/names"
)

// Named zz_* to run last: enabling the namespace check is final for every
// later test. An account whose owner no longer holds its namespace says so.
func TestOwnerOutsideNamespaceIsShown(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	Open(cross(cur), "zzlate")
	uassert.False(t, strings.Contains(Render("zzlate"), "Owner outside the namespace"))

	testing.SetRealm(testing.NewUserRealm("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx")) // r/sys/names admin
	names.Enable(cross(cur))
	uassert.True(t, strings.Contains(Render("zzlate"), "Owner outside the namespace"))

	// An account named after its owner's own address stays inside it.
	testing.SetRealm(testing.NewUserRealm(alice))
	Open(cross(cur), alice.String())
	uassert.False(t, strings.Contains(Render(alice.String()), "Owner outside the namespace"))
}
