// PKGPATH: gno.land/r/v3awalkterm
package v3awalkterm

import "chain/runtime"

// Phase B-min+ cornerstone: terminal frame.
//
// main() makes no nested call — it is the top of the call chain. From
// inside main's body the substrate has shifted m.Realm from nil to
// /r/v3awalkterm via Layer-1 (main is /r/-declared). The walk should
// surface this single transition:
//   - Self()         = /r/v3awalkterm (running realm)
//   - Caller()       = OriginCaller / EOA (no realm above main)
//   - caller.IsUser  = true
//   - Origin()       == Caller() at this depth

func main() {
	self := runtime.Self()
	caller := runtime.Caller()
	origin := runtime.Origin()

	println("self.PkgPath:    ", self.PkgPath())
	println("caller.IsUser:   ", caller.IsUser())
	println("caller==origin:  ", caller.Address() == origin.Address())
}

// Output:
// self.PkgPath:     gno.land/r/v3awalkterm
// caller.IsUser:    true
// caller==origin:   true
