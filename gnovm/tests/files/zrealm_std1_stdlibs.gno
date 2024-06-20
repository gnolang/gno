// PKGPATH: gno.land/r/std_test
package std_test

import (
	"std"
)

var aset *std.AddressList

func init() {
	caller := std.GetOrigCaller()
	aset = std.NewAddressList()
	aset.AddAddress(caller)
}

func main() {
	println(*aset)
	caller := std.GetOrigCaller()
	err := aset.AddAddress(caller)
	println("error:", err)
	has := aset.HasAddress(caller)
	println("has:", has)
	has = aset.HasAddress(std.Address(""))
	println("has:", has)
}

// Output:
// (slice[ref(1ed29bd278d735e20e296bd4afe927501941392f:5)] std.AddressList)
// error: address already exists
// has: true
// has: false
