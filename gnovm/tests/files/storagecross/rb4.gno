// PKGPATH: gno.land/r/caller
package caller

import (
	"gno.land/r/tests/vm/storagecross/alice"
	"gno.land/r/tests/vm/storagecross/bob"
)

func main() {

	// The storage realm(s) is initialy an empty stack (no methods have been called).
	//   - now: (storage:[], current:caller, previous:nil)

	// ================================================================================
	// ## fn is declared /r/realm:
	//
	// ----------------------------------------
	// CASE rB4: method has unattached receiver (var) -- do not storage-cross (same as rB3 above)
	//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
	//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
	obj := new(alice.Object)
	bob.Do(obj.Method) // FAIL: caller != bob
}

// Output:
// In bob.Do(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// In alice.Object.Method(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }

// Error:
// cannot modify external-realm or non-realm object
