// PKGPATH: gno.land/r/demo/groups
package groups

import (
	"chain/runtime"
	"testing"
)

func main(cur realm) {
	println("From main:")
	PrintRealm(cur)
	PrintRealm(cross)
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PreviousRealm:"), runtime.PreviousRealm())
	println("")

	println("From g1user origin:")
	testing.SetRealm(testing.NewUserRealm("g1user"))
	PrintRealm(cross)
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	func() {
		defer func() {
			r := recover()
			println(pad("PreviousRealm:"), r)
		}()
		println(pad("PreviousRealm:"), runtime.PreviousRealm())
	}()
	println("")

	println("From gno.land/r/sys/users realm:")
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
	PrintRealm(cross)
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PreviousRealm:"), runtime.PreviousRealm())
}

func pad(s string) string {
	for len(s) < 26 {
		s += " "
	}
	return s
}

func PrintRealm(cur realm) {
	println(pad("PrintRealm: CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PrintRealm: PreviousRealm:"), runtime.PreviousRealm())
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
