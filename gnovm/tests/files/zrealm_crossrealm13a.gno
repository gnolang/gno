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
// PrintRealm: CurrentRealm:  CodeRealm{ g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02, gno.land/r/demo/groups }
// PrintRealm: PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// PrintRealm: CurrentRealm:  CodeRealm{ g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02, gno.land/r/demo/groups }
// PrintRealm: PreviousRealm: CodeRealm{ g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02, gno.land/r/demo/groups }
// CurrentRealm:              CodeRealm{ g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02, gno.land/r/demo/groups }
// PreviousRealm:             UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
//
// From g1user origin:
// PrintRealm: CurrentRealm:  CodeRealm{ g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02, gno.land/r/demo/groups }
// PrintRealm: PreviousRealm: UserRealm{ g1user }
// CurrentRealm:              UserRealm{ g1user }
// PreviousRealm:             frame not found: cannot seek beyond origin caller override
//
// From gno.land/r/sys/users realm:
// PrintRealm: CurrentRealm:  CodeRealm{ g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02, gno.land/r/demo/groups }
// PrintRealm: PreviousRealm: CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users }
// CurrentRealm:              CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users }
// PreviousRealm:             CodeRealm{ g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02, gno.land/r/demo/groups }
