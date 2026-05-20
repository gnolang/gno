// PKGPATH: gno.land/r/v3acallernhost
package v3acallernhost

import "chain/runtime"

// runtime.CallerN exposes the height-N stack walk that v2's
// `rlm.Previous().Previous()...` chains relied on. Within the test
// harness (no cross-realm wrappers), CallerN(0)=Self, CallerN(1)=Caller
// (the EOA), and CallerN(2)+ would panic or hit the bounds of the
// available transitions.

func main() {
	self := runtime.Self()
	callerN0 := runtime.CallerN(0)
	callerN1 := runtime.CallerN(1)
	caller := runtime.Caller()

	// CallerN(0) == Self.
	println("CallerN(0)==Self addr:", callerN0.Address() == self.Address())
	println("CallerN(0)==Self path:", callerN0.PkgPath() == self.PkgPath())

	// CallerN(1) == Caller.
	println("CallerN(1)==Caller addr:", callerN1.Address() == caller.Address())
	println("CallerN(1)==Caller path:", callerN1.PkgPath() == caller.PkgPath())

	// CallerN(1) should be the EOA (empty pkgPath).
	println("CallerN(1) is user:", callerN1.IsUser())
}

// Output:
// CallerN(0)==Self addr: true
// CallerN(0)==Self path: true
// CallerN(1)==Caller addr: true
// CallerN(1)==Caller path: true
// CallerN(1) is user: true
