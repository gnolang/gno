// PKGPATH: gno.land/r/std_test
package std_test

import (
	"std"
)

// Like zrealm_std1.gno but interface var.
var aset std.AddressSet

func init() {
	// This won't work, init is called with empty caller.
	// caller := std.OriginCaller()
	caller := std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")
	aset = std.NewAddressList()
	aset.AddAddress(caller)
}

func main() {
	crossing()

	println(*(aset.(*std.AddressList)))
	caller := std.OriginCaller()
	err := aset.AddAddress(caller)
	println("error:", err)
	has := aset.HasAddress(caller)
	println("has:", has)
	has = aset.HasAddress(std.Address(""))
	println("has:", has)
}

// Output:
// (slice[ref(1ed29bd278d735e20e296bd4afe927501941392f:8)] std.AddressList)
// error: address already exists
// has: true
// has: false
