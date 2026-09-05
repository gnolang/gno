// PKGPATH: gno.land/r/gnoland/home/filetests/override_filetest

package override_filetest

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnoland/home"
)

func main(cur realm) {
	var admin = address("g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf")
	testing.SetOriginCaller(admin)
	home.AdminSetOverride(cross(cur), "Hello World!")
	println("---")
	println(home.Render(""))

	newAdmin := testutils.TestAddress("newAdmin")
	home.AdminTransferOwnership(cross(cur), newAdmin)
	if err := revive(func() {
		home.AdminSetOverride(cross(cur), "Not admin anymore")
	}); err == nil {
		panic("AdminSetOverride should have aborted the transaction")
	}
}

// Output:
// ---
// Hello World!
