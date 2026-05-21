// PKGPATH: gno.land/r/v3awalk3
package v3awalk3

import (
	"gno.land/r/tests/vm/v3aprobeb"
)

// Phase B-min+ cornerstone: 3-hop Layer-1 chain.
// /r/v3awalk3.main → /r/tests/vm/v3aprobeb.ChainProbe →
// /r/tests/vm/v3aprobe.Probe.
//
// All three hops are Layer-1 auto-cross. Inside Probe:
//   - Self()       = /r/tests/vm/v3aprobe
//   - Caller()     = /r/tests/vm/v3aprobeb
//   - ChainProbe's CallerN(2) inside /r/v3aprobeb = /r/v3awalk3

func main() {
	caller, self, prev := v3aprobeb.ChainProbe()
	println("caller.PkgPath:", caller.PkgPath())
	println("self.PkgPath:  ", self.PkgPath())
	println("prev.PkgPath:  ", prev.PkgPath())
}

// Output:
// caller.PkgPath: gno.land/r/tests/vm/v3aprobeb
// self.PkgPath:   gno.land/r/tests/vm/v3aprobe
// prev.PkgPath:   gno.land/r/v3awalk3
