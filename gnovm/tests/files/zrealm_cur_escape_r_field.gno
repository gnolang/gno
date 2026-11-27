// PKGPATH: gno.land/r/t73/escape_r_field
// A `cur` in a struct field: banker, Sub and cross refuse it once the call ended.
package escape_r_field

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
	probe(0, t73leak.LeakField(cross(cur)).R)
}

// Output:
// PkgPath: gno.land/r/tests/vm/t73leak
// Address: g1e8dt0awnnd3u47eq4p3zmqr6256vt370q38cqe
// IsCurrent: false
// banker refused: banker can only be instantiated for the current realm
// Sub banker refused: Sub: receiver is not the live cur (stale, sibling, or a sub-token)

// Error:
// cross: rlm is not the current cur (stale capture or sibling frame)
