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
	// CASE rA2: function declared in func(...) (non-crossing) -- do not storage-cross
	//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
	//   - in alice.TopFuncFnNoncrossing:       (storage:[], current:caller, previous:nil)
	//   - in alice.TopFuncFnNoncrossing.inner: (storage:[], current:caller, previous:nil)
	bob.Do(alice.TopFuncFnNoncrossing()) // FAIL: caller != bob

}

// Output:
// In alice.TopFuncFnNoncrossing(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// In bob.Do(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// In alice.TopFuncFnNoncrossing().inner, CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }

// Error:
// cannot modify external-realm or non-realm object
