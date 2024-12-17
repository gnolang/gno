package main

import (
	"std"
)

func main() {
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PrevRealm:"), std.PrevRealm())
	std.TestSetRealm(std.NewUserRealm("g1user"))
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PrevRealm:"), std.PrevRealm())
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/users"))
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PrevRealm:"), std.PrevRealm())
}

func pad(s string) string {
	for len(s) < 26 {
		s += " "
	}
	return s
}

func PrintRealm() {
	println(pad("PrintRealm: CurrentRealm:"), std.CurrentRealm())
	println(pad("PrintRealm: PrevRealm:"), std.PrevRealm())
}

// Because this is the context of a package, using PrintRealm()
// should not change the output of the main function.

// Output:
// PrintRealm: CurrentRealm:  (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: PrevRealm:     (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrevRealm:                 (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: CurrentRealm:  (struct{("g1user" std.Address),("" string)} std.Realm)
// PrintRealm: PrevRealm:     (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g1user" std.Address),("" string)} std.Realm)
// PrevRealm:                 (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: CurrentRealm:  (struct{("g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s" std.Address),("gno.land/r/demo/users" string)} std.Realm)
// PrintRealm: PrevRealm:     (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s" std.Address),("gno.land/r/demo/users" string)} std.Realm)
// PrevRealm:                 (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
