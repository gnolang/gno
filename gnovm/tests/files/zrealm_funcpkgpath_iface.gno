// PKGPATH: gno.land/r/demo/funcpkgiface
package funcpkgiface

import (
	"chain/runtime"

	plib "gno.land/p/demo/tests/funcpkg/v0"
)

type iface interface{ M() }

func main(cur realm) {
	var i iface = plib.NewT()
	f := i.M
	runtime.FuncPkgPath(f)
}

// Error:
// runtime.FuncPkgPath: method value of an interface has no fixed declaration
