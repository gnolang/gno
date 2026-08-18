// PKGPATH: gno.land/r/gnoland/home/filetests/override_filetest

package override_filetest

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnoland/home"
)

func main(cur realm) {
	var admin = address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh")
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
