// PKGPATH: gno.land/r/demo/funcpkguntypednil
package funcpkguntypednil

import "chain/runtime"

func main(cur realm) {
	runtime.FuncPkgPath(nil)
}

// Error:
// runtime.FuncPkgPath: not a function
