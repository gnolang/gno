// PKGPATH: gno.land/r/test/deferfixmixed
package deferfixmixed

// Test 6 (cross-realm variant): mix of deferred method calls (which
// need the receiver-aware borrow-switch fix) and deferred function
// calls (which already worked). Verifies LIFO order and that the
// IsBoundMethod flag is per-defer, not stack-state-leaking.

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type S struct {
	Tag string
}

// Mutates s.Tag — exercises the borrow-switch for cross-realm receiver.
func (s *S) Append(suffix string) {
	s.Tag = s.Tag + suffix
	println("method appended; tag now:", s.Tag)
}

func bareFunc() {
	println("function: bare")
}

func main(cur realm) {
	s := &S{Tag: "x"}
	crossrealm.SetObject(cross, s)

	defer bareFunc()       // 1st registered, 3rd to fire (LIFO)
	defer s.Append("-mid") // 2nd registered, 2nd to fire — needs borrow-switch
	defer bareFunc()       // 3rd registered, 1st to fire
	println("body done")
}

// Output:
// body done
// function: bare
// method appended; tag now: x-mid
// function: bare
