// PKGPATH: gno.land/r/test/methodexprinline
package methodexprinline

// Probe: `Type.Method(recv)` (method expression) on a cross-realm-stored
// receiver. This dispatches as an ordinary FuncValue call with `recv` as
// args[0] — not as a BoundMethodValue — so neither the N_Readonly strip
// nor the implicit borrow-realm switch in PushFrameCall fires. Expected
// to panic with the realm-ownership-mismatch ("readonly tainted") message
// because m.Realm stays as the caller realm.

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type S struct {
	Closed bool
}

func (s *S) Close() {
	s.Closed = true
}

func main(cur realm) {
	s := &S{}
	crossrealm.SetObject(cross, s)
	(*S).Close(s) // inline method-expression call
	println("should not print")
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): s<VPBlock(1,0)>.Closed
