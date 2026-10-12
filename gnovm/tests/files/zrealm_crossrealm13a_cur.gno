// PKGPATH: gno.land/r/demo/groups
package groups

import (
	"testing"
)

// Exercises the captured `cur realm` handle (uverse interface) against
// unsafe.CurrentRealm() / unsafe.PreviousRealm() (chain/runtime struct),
// including under successive testing.SetRealm overrides — both APIs
// produce identical output by design.

func main(cur realm) {
	println("From main:")
	PrintRealm(cur)
	PrintRealm(cross(cur))
	println(pad("CurrentRealm:"), renderCur(cur))
	println(pad("PreviousRealm:"), renderPrev(cur))
	println("")

	println("From g1user origin:")
	testing.SetRealm(testing.NewUserRealm("g1user"))
	PrintRealm(cross(cur))
	println(pad("CurrentRealm:"), renderCur(cur))
	println(pad("PreviousRealm:"), renderPrev(cur))
	println("")

	println("From gno.land/r/sys/users realm:")
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
	PrintRealm(cross(cur))
	println(pad("CurrentRealm:"), renderCur(cur))
	println(pad("PreviousRealm:"), renderPrev(cur))
}

func pad(s string) string {
	for len(s) < 26 {
		s += " "
	}
	return s
}

// format renders an addr+pkgpath pair in the same format
// runtime.Realm.String() uses: "CodeRealm{ <addr>, <pkgpath> }" when
// the pkgpath is non-empty, "UserRealm{ <addr> }" when it isn't.
func format(addr, path string) string {
	if path == "" {
		return "UserRealm{ " + addr + " }"
	}
	return "CodeRealm{ " + addr + ", " + path + " }"
}

// renderCur / renderPrev are thin helpers (avoid taking `realm` as a
// parameter so they aren't crossing-functions — that would require
// callers to use `cross`/`cur` literally as args).
func renderCur(cur realm) string {
	return format(string(cur.Address()), cur.PkgPath())
}

// renderPrev recovers cur.Previous()'s walk-end panic internally and
// returns the panic string so callers don't need a defer-recover wrapper
// (which would require closing over `cur`, forbidden by Gno).
func renderPrev(cur realm) (out string) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				out = s
			} else {
				out = "<recovered>"
			}
		}
	}()
	prev := cur.Previous()
	if prev == nil {
		return "<nil>"
	}
	return format(string(prev.Address()), prev.PkgPath())
}

func PrintRealm(cur realm) {
	println(pad("PrintRealm: CurrentRealm:"), renderCur(cur))
	println(pad("PrintRealm: PreviousRealm:"), renderPrev(cur))
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
