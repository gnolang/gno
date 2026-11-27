// PKGPATH: gno.land/r/t73/escape_r_closure_save
// A closure that captured `cur`, saved to a package variable: refused.
package escape_r_closure_save

import (
	"chain/banker"

	"gno.land/r/tests/vm/t73leak"
)

func try(what string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			println(what, "refused:", r)
		}
	}()
	f()
	println(what, "allowed")
}

// probe shows what a holder of another realm's stale `cur` can do;
// a refused cross aborts, so it comes last.
func probe(_ int, r realm) {
	println("PkgPath:", r.PkgPath())
	println("Address:", r.Address())
	println("IsCurrent:", r.IsCurrent())
	try("banker", func() { banker.NewBanker(banker.BankerTypeRealmSend, r) })
	try("Sub banker", func() { banker.NewBanker(banker.BankerTypeRealmSend, r.Sub("x")) })
	t73leak.Bump(cross(r))
	println("not reached")
}

func main(cur realm) {
	saved = t73leak.LeakClosure(cross(cur))
	println("stored; refused when the realm is saved")
}

var saved func(string)

// Output:
// stored; refused when the realm is saved

// Error:
// cannot persist realm value: realm values are ephemeral and tied to a call frame
