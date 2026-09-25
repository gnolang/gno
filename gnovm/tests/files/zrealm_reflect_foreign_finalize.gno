// PKGPATH: gno.land/r/demo/reflectcross
package reflectcross

// An object belongs to the realm that created it, even when a different realm
// is the one that persists it.
//
// `Realm.assignNewObjectID` takes the tick from the object's own PkgID realm,
// through `touchForeignRealm`, not from the realm running the finalize. That is
// the branch this exercises, and it is the one that would break the whole
// scheme if it went the other way: two realms minting from one counter would
// hand out the creating realm's tick twice, and two objects would share an
// address.

import (
	"chain/reflect"

	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type fooer struct{ n int }

func (fooer) Foo(cur realm) {}
func (fooer) Bar()          {}

var (
	// mine is created and persisted by this realm, the ordinary case.
	mine = &fooer{n: 1}
	// theirs is created here and handed to another realm, which is what
	// persists it.
	theirs = &fooer{n: 2}
)

func main(cur realm) {
	// crossrealm stores `theirs` in its own tree, so crossrealm's finalize is
	// what stamps it.
	crossrealm.SetFooer(cross(cur), theirs)

	m, mok := reflect.Of(mine)
	t, tok := reflect.Of(theirs)
	println("both stamped:", mok && tok)

	// The creating realm is this one for both, not the realm that persisted
	// the second.
	println("mine   created by:", m.PkgPath())
	println("theirs created by:", t.PkgPath())
	println("same creating realm:", m.PkgPath() == t.PkgPath())

	// And the two took different ticks of that one realm's clock, so the
	// foreign finalize did not reuse a tick this realm had already spent.
	println("addresses differ:", m.Address() != t.Address())
	println("ids differ:      ", m.ID() != t.ID())
}

// Output:
// both stamped: true
// mine   created by: gno.land/r/demo/reflectcross
// theirs created by: gno.land/r/demo/reflectcross
// same creating realm: true
// addresses differ: true
// ids differ:       true
