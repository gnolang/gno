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
	// CASE rC5: method has attached receiver via cls late -- do not storage-cross
	//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
	//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
	obj := new(alice.Object)
	var cls func()
	cls = func() {
		// bob.Do(obj.Method) // FAIL: not yet unattached
		bob.SetClosure(cls) // obj recursively attached to alice late
		bob.Do(obj.Method)  // SUCCESS: attached to bob
	}
	cls()
	// FAILED:
	// cannot modify external-realm or non-realm object
	// XXX These parts of the code are not reached.
	// In bob.Do(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
	// In alice.Object.Method(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }

}

// Output:
// In bob.Do(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// In alice.Object.Method(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
