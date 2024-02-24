// PKGPATH: gno.land/r/test
package test

import (
	"std"

	"gno.land/p/demo/access"
	"gno.land/p/demo/testutils"
)

var Admins = access.New().Unprivileged()

func main() {
	initialCaller := std.GetOrigCaller()
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")
	println(Admins.List())
	Admins.Add(test1)
	println(Admins.List())
	Admins.Add(test1)
	println(Admins.List())
	println(Admins.CurrentHasAccess())
	Admins.Del(initialCaller)
	println(Admins.List())
	println(Admins.CurrentHasAccess())
	std.TestSetOrigCaller(test2)
	println(Admins.CurrentHasAccess())
	std.TestSetOrigCaller(test1)
	println(Admins.CurrentHasAccess())
}

// Output:
// slice[ref(a8ada09dee16d791fd406d629fe29bb0ed084a30:6)]
// slice[("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7" std.Address)]
// slice[("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7" std.Address)]
// true
// slice[("g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7" std.Address)]
// false
// false
// true
