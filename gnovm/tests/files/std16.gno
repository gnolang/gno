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

// A panic that crosses a realm boundary is an abort, so revive, not recover.
func aborted(f func()) bool {
	return revive(f) != nil
}

func main(cur realm) {
	// Same-realm shapes: a callback runner and a named wrapper in this realm
	// are this realm's own code, so neither is a call boundary.
	callFunc(func() { Register() })
	println("callback to own named func ok")

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
// callback to own named func ok
// callback through another realm's named func panicked: true
// second realm panicked: true
// direct realm call ok
