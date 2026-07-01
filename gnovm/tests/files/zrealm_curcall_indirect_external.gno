// PKGPATH: gno.land/r/test/indirectcurexternal
package indirectcurexternal

// Companion to z_indirect_cur_call_local: the fix defers the
// same-realm check to runtime when the function is held in a local
// variable. Verify the runtime still rejects a cross-realm cur-call.

import tests "gno.land/r/tests/vm"

func main(cur realm) {
	p := tests.IncCounter // *FuncValue from a different realm
	p(cur)             // cur-call to external realm via local var
}

// Error:
// cannot cur-call to external realm function gno.land/r/tests/vm.IncCounter from gno.land/r/test/indirectcurexternal
