// PKGPATH: gno.land/r/caller
package caller

import (
	"gno.land/r/tests/vm/storagecross/bob"
)

func main() {

	// The storage realm(s) is initialy an empty stack (no methods have been called).
	//   - now: (storage:[], current:caller, previous:nil)

	// ================================================================================
	// ## fn is declared /r/realm:
	//
	// --------------------------------------------------------------------------------
	// CASE rB1: method has nil receiver -- storage-cross to declared
	//   - in bob.Do:                           (storage:[],    current:caller, previous:nil)
	//   - in obj.Method:                       (storage:[bob], current:caller, previous:nil)
	bob.Do((*bob.Object)(nil).Method) // SUCCESS: storage-crossed to bob
	// FAILED:
	// unexpected panic: cannot modify external-realm or non-realm object

}

// Output:
// In bob.Do(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// In bob.Object.Method(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
