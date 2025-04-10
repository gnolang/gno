// PKGPATH: gno.land/r/demo/groups
package groups

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
// PrintRealm: PreviousRealm: (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" std.Address),("gno.land/r/sys/users" string)} std.Realm)
// CurrentRealm:              (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" std.Address),("gno.land/r/sys/users" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
