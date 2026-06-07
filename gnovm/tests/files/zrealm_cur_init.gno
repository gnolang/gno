// PKGPATH: gno.land/r/curinit
package curinit

import (
	"chain/runtime/unsafe"
)

// init(cur realm) is entered as a non-cross call via the .cur ConstExpr
// path (same machinery as main(cur realm)). The doOpCall fix detects the
// preprocess-time placeholder origin and rebuilds with the per-tx origin
// so init's cur.Previous() reflects the deployer EOA — mirroring
// unsafe.PreviousRealm() for StageAdd height=1.
var initCurPkgPath, initPrevPkgPath string

func init(cur realm) {
	initCurPkgPath = cur.PkgPath()
	initPrevPkgPath = cur.Previous().PkgPath()
}

func main(cur realm) {
	println("init cur.PkgPath:", initCurPkgPath)
	println("init prev.PkgPath empty (EOA):", initPrevPkgPath == "")
	println("main cur.PkgPath:", cur.PkgPath())
	println("main prev.PkgPath empty (EOA):", cur.Previous().PkgPath() == "")
	println("runtime.Current.PkgPath:", unsafe.CurrentRealm().PkgPath())
}

// Output:
// init cur.PkgPath: gno.land/r/curinit
// init prev.PkgPath empty (EOA): true
// main cur.PkgPath: gno.land/r/curinit
// main prev.PkgPath empty (EOA): true
// runtime.Current.PkgPath: gno.land/r/curinit
