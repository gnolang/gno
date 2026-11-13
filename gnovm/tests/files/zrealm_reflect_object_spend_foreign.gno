// PKGPATH: gno.land/r/demo/objspendforeign
package objspendforeign

// The authority rule, from the other side: a realm cannot mint a banker for an
// object it does not own, even when it is holding the object and can read its
// address.
//
// This is what makes an object address an account rather than a label. Holding
// a reference to an object is not authority over its funds, and nothing a
// realm can write down changes that: the check is an equality between the
// PkgID inside the object's VM identity and the hash of the calling realm's
// own path, both computed by the VM.

import (
	"chain/banker"
	"chain/reflect"

	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type mine struct{ n int }

var ours = &mine{1}

func main(cur realm) {
	// Our own object: a banker for it is fine, and the realm can spend.
	own, ok := reflect.Of(ours)
	println("own object stamped:", ok)
	println("own object is ours: ", own.PkgPath() == "gno.land/r/demo/objspendforeign")
	banker.NewObjectBanker(own, cur)
	println("banker for our own object: ok")

	// crossrealm's `local` is created in crossrealm's init, so crossrealm's
	// clock minted it. We can reach it, read it, and derive its address.
	theirs, ok := reflect.Of(crossrealm.Make1().B)
	println("their object stamped:", ok)
	println("their object is theirs:", theirs.PkgPath() == "gno.land/r/tests/vm/crossrealm")
	println("we can read their address:", theirs.Address() != "")

	// Reading it is all we can do.
	banker.NewObjectBanker(theirs, cur)
	println("unreachable")
}

// Output:
// own object stamped: true
// own object is ours:  true
// banker for our own object: ok
// their object stamped: true
// their object is theirs: true
// we can read their address: true

// Error:
// object "07e973c6aeab1f1cb747d0903cb1e4d369226376:45" was not minted by "gno.land/r/demo/objspendforeign": only the realm whose clock minted an object can spend from its address
