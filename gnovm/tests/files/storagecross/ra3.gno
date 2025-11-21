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
	// CASE rA3: function declared in func(cur realm, ...) (crossing) -- do not storage-cross
	//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
	//   - alice.TopFuncFnCrossing:             (storage:[], current:alice,  previous:caller)
	//   - alice.TopFuncFnCrossing.inner:       (storage:[], current:caller, previous:nil)
	//     (cross doesn't matter in inner)
	bob.Do(alice.TopFuncFnCrossing(cross)) // FAIL: caller != bob

}

// Output:
// In alice.TopFuncFnCrossing(), CurrentRealm: gno.land/r/tests/vm/storagecross/alice PreviousRealm: CodeRealm{ g1t2l24p70a39k2c9pednxmlc6e33hrlz74p800p, gno.land/r/caller }
// In bob.Do(), CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// In alice.TopFuncFnCrossing().inner, CurrentRealm: gno.land/r/caller PreviousRealm: UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }

// Error:
// cannot modify external-realm or non-realm object
