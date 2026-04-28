// PKGPATH: gno.land/r/test/deferfixrecover
package deferfixrecover

// Test 7: Security regression. Pre-fix, an adversarial granter could
// install an outer defer-recover wrapping `defer al.Close()` to swallow
// the readonly-taint panic raised by deferred-method on a cross-realm-
// stored struct. The mutation never happened, so a kill-switch flag
// stayed unset across tx boundaries — security broken.
//
// Post-fix, the deferred method call performs the implicit borrow-realm
// switch, the mutation completes cleanly, and the outer recover finds
// nothing to swallow. The flag is correctly set.

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type Switch struct {
	Closed bool
}

func (s *Switch) Close() {
	s.Closed = true
}

func main(cur realm) {
	defer func() {
		if r := recover(); r != nil {
			println("RECOVERED:", r)
		}
	}()

	s := &Switch{}
	crossrealm.SetObject(cross, s)

	defer s.Close()

	println("Closed before defer fires:", s.Closed)
}

// Output:
// Closed before defer fires: false
