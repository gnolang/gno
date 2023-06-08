// PKGPATH: gno.land/r/test
package test

import (
	"std"

	"gno.land/p/demo/administrable"
	"gno.land/p/demo/testutils"
)

var admins administrable.Set

func init() {
	admins = administrable.New()
}

func main() {
	initialCaller := std.GetOrigCaller()
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")
	println(admins.List())
	admins.Add(test1)
	println(admins.List())
	admins.Add(test1)
	println(admins.List())
	println(admins.IsAdmin())
	admins.Del(initialCaller)
	println(admins.List())
	println(admins.IsAdmin())
	std.TestSetOrigCaller(test2)
	println(admins.IsAdmin())
	std.TestSetOrigCaller(test1)
	println(admins.IsAdmin())
}

// Output:
// slice[ref(a8ada09dee16d791fd406d629fe29bb0ed084a30:5)]
// slice[("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7" std.Address)]
// slice[("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7" std.Address)]
// true
// slice[("g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7" std.Address)]
// false
// false
// true
