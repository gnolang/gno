package main

import (
	"std"
	"testing"
)

func main() {
	// Need this as base frame because main2 calls PreviousRealm.
	testing.SetRealm(std.NewUserRealm("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	main2()
}

func main2() {
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PreviousRealm:"), std.PreviousRealm())
	testing.SetRealm(std.NewUserRealm("g1user"))
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PreviousRealm:"), std.PreviousRealm())
	testing.SetRealm(std.NewCodeRealm("gno.land/r/sys/users"))
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PreviousRealm:"), std.PreviousRealm())
}

func pad(s string) string {
	for len(s) < 26 {
		s += " "
	}
	return s
}

func PrintRealm() {
	println(pad("PrintRealm: CurrentRealm:"), std.CurrentRealm())
	println(pad("PrintRealm: PreviousRealm:"), std.PreviousRealm())
}

// Output:
// PrintRealm: CurrentRealm:  (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: CurrentRealm:  (struct{("g1user" std.Address),("" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g1user" std.Address),("" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: CurrentRealm:  (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" std.Address),("gno.land/r/sys/users" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" std.Address),("gno.land/r/sys/users" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
