package wugnot

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/testutils"
	"github.com/gnolang/gno/gnovm/tests/stdlibs/std"
)

var (
	alice = testutils.TestAddress("alice") // g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
	bob   = testutils.TestAddress("bob")   // g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
)

func TestDeposit() {
	/// how do we get an initial balance of ugnot to a test address? without using test1?

	std.TestSetOrigCaller(alice)

}
