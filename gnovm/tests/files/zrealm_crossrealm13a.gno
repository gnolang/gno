// PKGPATH: gno.land/r/demo/groups
package groups

import (
	"std"
)

func main() {
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PreviousRealm:"), std.PreviousRealm())
	std.TestSetRealm(std.NewUserRealm("g1user"))
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PreviousRealm:"), std.PreviousRealm())
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/users"))
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
// PrintRealm: CurrentRealm:  (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: CurrentRealm:  (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1user" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g1user" std.Address),("" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: CurrentRealm:  (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s" std.Address),("gno.land/r/demo/users" string)} std.Realm)
// CurrentRealm:              (struct{("g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s" std.Address),("gno.land/r/demo/users" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
