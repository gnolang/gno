// PKGPATH: gno.land/r/test/methodexprdefer
package methodexprdefer

// Probe: `defer Type.Method(recv)` on a cross-realm-stored receiver.
// Because this is a deferred FuncValue call (not a deferred bound method),
// dfr.IsBoundMethod stays false, the deferred-call path in
// doOpReturnCallDefers passes TypedValue{} for the receiver, and
// PushFrameCall does not perform the implicit borrow-realm switch.
// The defer-method fix does not apply to this form. Expected to panic.

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
	defer (*S).Close(s)
	println("body done")
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): s<VPBlock(1,0)>.Closed
