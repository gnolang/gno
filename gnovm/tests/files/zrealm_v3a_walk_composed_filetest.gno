// PKGPATH: gno.land/r/v3awalkcomp
package v3awalkcomp

import (
	"gno.land/p/nt/v3acomposed"
	"gno.land/r/tests/vm/v3aprobe"
)

// Phase B-min+ cornerstone: /p/ transparent frame then Layer-1.
//
// Call chain: /r/v3awalkcomp → /p/nt/v3acomposed.Run → fn() where
// fn is the /r/-declared v3aprobe.Probe function value.
//
// The /p/ frame is transparent: m.Realm doesn't shift on entry to
// Run (no Layer-1 — /p/ is not a realm; no Layer-2 — no receiver),
// so m.Realm stays /r/v3awalkcomp through Run's body. When fn()
// invokes Probe, Layer-1 shifts to /r/v3aprobe. The walk should
// report Caller inside Probe as /r/v3awalkcomp, NOT /p/v3acomposed,
// because the substrate never shifted to /p/.

func main() {
	caller, self := v3acomposed.Run(v3aprobe.Probe)
	println("caller.PkgPath:", caller.PkgPath())
	println("self.PkgPath:  ", self.PkgPath())
}

// Output:
// caller.PkgPath: gno.land/r/v3awalkcomp
// self.PkgPath:   gno.land/r/tests/vm/v3aprobe
