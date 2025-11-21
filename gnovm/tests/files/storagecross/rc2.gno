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
	// CASE rC2: method has (wrong) attached receiver -- storage-cross to receiver
	//  - in bob.Do:                            (storage:[],    current:caller, previous:nil)
	//  - in obj.Method:                        (storage:[alice], current:caller, previous:nil)
	obj := new(alice.Object)
	alice.SetObject(obj) // obj attached to alice
	bob.Do(obj.Method)   // FAIL: attached to alice (but need bob)
	//
	// XXX: These parts of the code are not reached.
	// In bob.Do(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
	// In alice.Object.Method(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }

}

// Error:
// cannot modify external-realm or non-realm object
