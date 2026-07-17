// PKGPATH: gno.land/r/test/indirectcurexternal
package indirectcurexternal

// Companion to zrealm_curcall_indirect_local.gno. When the called function is
// held in a local variable, preprocess can't statically bind it, so the
// same-realm check defers to runtime. Verify the runtime still rejects a
// cross-realm cur-call. The absence of a file:line prefix on the expected
// error confirms it comes from machine.go, not preprocess.go.

import tests "gno.land/r/tests/vm"

func main(cur realm) {
	p := tests.IncCounter // *FuncValue from a different realm
	p(cur)                // cur-call to external realm via local var
}

// Error:
// cannot cur-call to external realm function gno.land/r/tests/vm.IncCounter from gno.land/r/test/indirectcurexternal
