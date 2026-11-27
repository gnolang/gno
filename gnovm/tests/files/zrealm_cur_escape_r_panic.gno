// PKGPATH: gno.land/r/t73/escape_r_panic
// A `cur` carried by a panic out of a same-realm cur-call, recovered by the
// caller: the value is the caller's own live `cur`. (A panic out of a
// `cross` call is not recoverable, so it cannot carry `cur` out.)
package escape_r_panic

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

// probe shows what the holder of the recovered value can do.
func probe(_ int, r realm) {
	println("PkgPath:", r.PkgPath())
	println("Address:", r.Address())
	println("IsCurrent:", r.IsCurrent())
	try("banker", func() { banker.NewBanker(banker.BankerTypeRealmSend, r) })
	try("Sub banker", func() { banker.NewBanker(banker.BankerTypeRealmSend, r.Sub("x")) })
	t73leak.Bump(cross(r))
	println("Count:", t73leak.Count)
}

// leakPanic is a crossing function of this realm: its `cur` is its own.
func leakPanic(cur realm) { panic(cur) }

func main(cur realm) {
	var held realm
	func() {
		defer func() {
			r := recover()
			println("recovered:", r != nil)
			held, _ = r.(realm)
		}()
		leakPanic(cur)
	}()
	println("held is caller cur:", held == cur)
	probe(0, held)
}

// Output:
// recovered: true
// held is caller cur: true
// PkgPath: gno.land/r/t73/escape_r_panic
// Address: g16wy6q3dn7srcjmfzr7vx5ktk4ryqv8hgm7345s
// IsCurrent: true
// banker allowed
// Sub banker allowed
// Count: 1
