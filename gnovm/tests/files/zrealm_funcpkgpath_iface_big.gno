// MAXALLOC: 100000000
// PKGPATH: gno.land/r/demo/funcpkgifacebig
package funcpkgifacebig

import "chain/runtime"

type iface interface{ M() }

type Big struct{ a [60000000]byte }

func (Big) M() {}

func main(cur realm) {
	p := &Big{}
	var i iface = p
	f := i.M
	runtime.FuncPkgPath(f)
}

// Error:
// runtime.FuncPkgPath: method value of an interface has no fixed declaration
