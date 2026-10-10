// PKGPATH: gno.land/r/test
package test

import (
	"chain/runtime"

	rtests "gno.land/r/tests/vm"
	rsubtests "gno.land/r/tests/vm/subtests"
)

func Register() {
	runtime.AssertOriginCall()
}

func callFunc(f func()) {
	f()
}

// main is the message's own script, so r/test is never an intermediary: a
// direct cross into r/tests/vm is an origin call, while any r/test frame
// between main and the cross (aborted, callFunc) enters r/test first and
// makes it the entry.
//
// A panic that crosses a realm boundary is an abort, so revive, not recover.
func aborted(f func()) bool {
	return revive(f) != nil
}

func panicked(f func()) (r any) {
	defer func() { r = recover() }()
	f()
	return nil
}

func main(cur realm) {
	// Same-realm shape: callFunc is the entry, the closure is transparent, but
	// Register is a named function of this realm above the entry, a value a
	// script could hand back, so it is a boundary. Same realm, so recover.
	println("callback to own named func panicked:",
		panicked(func() { callFunc(func() { Register() }) }))

	// Security case 1: the callback runner belongs to ANOTHER realm. The
	// closure is ours, but it is called by subtests, so subtests is on the
	// stack between us and the assertion.
	println("callback through another realm's named func panicked:",
		aborted(func() { rsubtests.Exec(func() { Register() }) }))

	// Security case 2: two realms entered in a row.
	println("second realm panicked:",
		aborted(func() { rtests.CallSubtestsAssertOriginCall(cross(cur)) }))

	// Sanity: entering one realm directly is an origin call.
	rtests.CallAssertOriginCall(cross(cur))
	println("direct realm call ok")
}

// Output:
// callback to own named func panicked: invalid non-origin call
// callback through another realm's named func panicked: true
// second realm panicked: true
// direct realm call ok
