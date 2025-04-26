// PKGPATH: gno.land/r/demo/groups
package groups

import (
	"std"
	"testing"
)

func main() {
	crossing()

	println("From main:")
	PrintRealm()
	cross(PrintRealm)()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PreviousRealm:"), std.PreviousRealm())
	println("")

	println("From g1user origin:")
	testing.SetRealm(std.NewUserRealm("g1user"))
	cross(PrintRealm)()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	func() {
		defer func() {
			r := recover()
			println(pad("PreviousRealm:"), r)
		}()
		println(pad("PreviousRealm:"), std.PreviousRealm())
	}()
	println("")

	println("From gno.land/r/sys/users realm:")
	testing.SetRealm(std.NewCodeRealm("gno.land/r/sys/users"))
	cross(PrintRealm)()
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
	crossing()

	println(pad("PrintRealm: CurrentRealm:"), std.CurrentRealm())
	println(pad("PrintRealm: PreviousRealm:"), std.PreviousRealm())
}

// Output:
// From main:
// PrintRealm: CurrentRealm:  (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// PrintRealm: CurrentRealm:  (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// CurrentRealm:              (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
//
// From g1user origin:
// PrintRealm: CurrentRealm:  (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1user" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g1user" std.Address),("" string)} std.Realm)
// PreviousRealm:             frame not found: cannot seek beyond origin caller override
//
// From gno.land/r/sys/users realm:
// PrintRealm: CurrentRealm:  (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
// PrintRealm: PreviousRealm: (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" std.Address),("gno.land/r/sys/users" string)} std.Realm)
// CurrentRealm:              (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" std.Address),("gno.land/r/sys/users" string)} std.Realm)
// PreviousRealm:             (struct{("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02" std.Address),("gno.land/r/demo/groups" string)} std.Realm)
