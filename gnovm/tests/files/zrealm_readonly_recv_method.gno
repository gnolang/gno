// PKGPATH: gno.land/r/test/readonlyrecvmethod
package readonlyrecvmethod

// Test B: same setup as Test A — pointer with sticky N_Readonly from
// an external /p/ package — but call a method that mutates instead
// of writing the field directly. Under the pre-interrealm-v2-Phase-3
// design, method dispatch laundered the N_Readonly taint and the
// mutation succeeded silently. Under Phase 3 (persistent frozen /p/
// realms), the borrow rule routes m.Realm to /p/demo/tests's frozen
// Realm during Modify()'s body; the direct write of `o2.Field` reaches
// DidUpdate with a Frozen realm and panics. This closes the launder.

import (
	"gno.land/p/demo/tests"
)

func main(cur realm) {
	p := tests.SomeValue3 // *TestRealmObject2 — sticky N_Readonly propagates
	p.Modify()            // o2.Field = "modified" — does the receiver's bit survive?
	println("after Modify; Field =", p.Field)
}

// Error:
// cannot mutate immutable package gno.land/p/demo/tests: object 4609759cdf9e3b49753fd5632a02365700aff2b6:21
