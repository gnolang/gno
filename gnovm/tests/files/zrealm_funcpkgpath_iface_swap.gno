// PKGPATH: gno.land/r/demo/funcpkgifaceswap
package funcpkgifaceswap

import (
	"chain/runtime"

	plib "gno.land/p/demo/tests/funcpkg/v0"
)

type iface interface{ M() }

type box struct{ iface }

type evil struct{}

func (evil) M() {}

func probe(name string, f func()) {
	defer func() {
		println(name, "panicked:", recover())
	}()
	println(name, "=>", runtime.FuncPkgPath(f))
}

func main(cur realm) {
	b := &box{plib.NewT()}
	var i iface = b
	f := i.M // one value, never reassigned
	probe("lazy before swap", f)
	b.iface = evil{}
	probe("lazy after swap", f)
	// b.M is promoted through the interface field, so it is a lazy bind too.
	probe("promoted b.M", b.M)
	pt := plib.NewT()
	g := pt.M // eagerly bound: the concrete method is fixed when bound
	println("eager before swap:", runtime.FuncPkgPath(g))
	b.iface = plib.NewT()
	println("eager after swap:", runtime.FuncPkgPath(g))
}

// Output:
// lazy before swap panicked: runtime.FuncPkgPath: method value of an interface has no fixed declaration
// lazy after swap panicked: runtime.FuncPkgPath: method value of an interface has no fixed declaration
// promoted b.M panicked: runtime.FuncPkgPath: method value of an interface has no fixed declaration
// eager before swap: gno.land/p/demo/tests/funcpkg/v0
// eager after swap: gno.land/p/demo/tests/funcpkg/v0
