// PKGPATH: gno.land/r/t73/escape_r_closure
// A closure that captured `cur` uses it after the call ended: banker, Sub and cross refuse it.
package escape_r_closure

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
	f := t73leak.LeakClosure(cross(cur))
	f("PkgPath")
	try("banker", func() { f("banker") })
	try("Sub banker", func() { f("Sub") })
	f("cross")
	println("not reached")
}

// Output:
// closure PkgPath: gno.land/r/tests/vm/t73leak
// banker refused: banker can only be instantiated for the current realm
// Sub banker refused: Sub: receiver is not the live cur (stale, sibling, or a sub-token)

// Error:
// cross: rlm is not the current cur (stale capture or sibling frame)
