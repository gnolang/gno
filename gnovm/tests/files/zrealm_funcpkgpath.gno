// PKGPATH: gno.land/r/demo/funcpkgcaller
package funcpkgcaller

import (
	"chain/runtime"

	plib "gno.land/p/demo/tests/funcpkg/v0"
	funcpkg "gno.land/r/tests/vm/funcpkg"
)

type callerType struct{}

func (callerType) M() {}

type handler func()

func ownTop() {}

func main(cur realm) {
	otherFn := funcpkg.GetFn(cross(cur))
	println("other realm top-level: " + runtime.FuncPkgPath(otherFn))
	println("own top-level: " + runtime.FuncPkgPath(ownTop))

	lit := func() {}
	println("own literal: " + runtime.FuncPkgPath(lit))

	println("p top-level: " + runtime.FuncPkgPath(plib.Top))
	println("p-built closure: " + runtime.FuncPkgPath(plib.BuildClosure()))

	var ct callerType
	println("realm type method: " + runtime.FuncPkgPath(ct.M))

	pt := plib.NewT()
	println("p type method: " + runtime.FuncPkgPath(pt.M))

	var h handler = ownTop
	println("named func type: " + runtime.FuncPkgPath(h))

	var typedNil func()
	println("typed nil: \"" + runtime.FuncPkgPath(typedNil) + "\"")
}

// Output:
// other realm top-level: gno.land/r/tests/vm/funcpkg
// own top-level: gno.land/r/demo/funcpkgcaller
// own literal: gno.land/r/demo/funcpkgcaller
// p top-level: gno.land/p/demo/tests/funcpkg/v0
// p-built closure: gno.land/p/demo/tests/funcpkg/v0
// realm type method: gno.land/r/demo/funcpkgcaller
// p type method: gno.land/p/demo/tests/funcpkg/v0
// named func type: gno.land/r/demo/funcpkgcaller
// typed nil: ""
