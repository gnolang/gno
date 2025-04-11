package main

import (
	"chain/runtime"
	"testing"
)

func main() {
	PrintRealm()
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PreviousRealm:"), runtime.PreviousRealm())
	testing.SetRealm(runtime.NewUserRealm("g1user"))
	PrintRealm()
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PreviousRealm:"), runtime.PreviousRealm())
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/sys/users"))
	PrintRealm()
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PreviousRealm:"), runtime.PreviousRealm())
}

func pad(s string) string {
	for len(s) < 26 {
		s += " "
	}
	return s
}

func PrintRealm() {
	println(pad("PrintRealm: CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PrintRealm: PreviousRealm:"), runtime.PreviousRealm())
}

// Because this is the context of a package, using PrintRealm()
// should not change the output of the main function.

// Output:
// PrintRealm: CurrentRealm:  (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" chain.Address),("" string)} chain/runtime.Realm)
// PrintRealm: PreviousRealm: (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" chain.Address),("" string)} chain/runtime.Realm)
// CurrentRealm:              (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" chain.Address),("" string)} chain/runtime.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" chain.Address),("" string)} chain/runtime.Realm)
// PrintRealm: CurrentRealm:  (struct{("g1user" chain.Address),("" string)} chain/runtime.Realm)
// PrintRealm: PreviousRealm: (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" chain.Address),("" string)} chain/runtime.Realm)
// CurrentRealm:              (struct{("g1user" chain.Address),("" string)} chain/runtime.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" chain.Address),("" string)} chain/runtime.Realm)
// PrintRealm: CurrentRealm:  (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" chain.Address),("gno.land/r/sys/users" string)} chain/runtime.Realm)
// PrintRealm: PreviousRealm: (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" chain.Address),("" string)} chain/runtime.Realm)
// CurrentRealm:              (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" chain.Address),("gno.land/r/sys/users" string)} chain/runtime.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" chain.Address),("" string)} chain/runtime.Realm)
