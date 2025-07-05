// PKGPATH: gno.land/r/std
package std

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

func main(cur realm) {
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
// (slice[ref(6da16ec68c07a7e26176ab66972e2ddf979cd22f:8)] std.AddressList)
// error: address already exists
// has: true
// has: false
