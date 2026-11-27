// PKGPATH: gno.land/r/demo/funcpkgnotfunc
package funcpkgnotfunc

import "chain/runtime"

func main(cur realm) {
	n := 42
	runtime.FuncPkgPath(n)
}

// Error:
// runtime.FuncPkgPath: not a function
