// PKGPATH: gno.land/r/curpersistclosure
package curpersistclosure

// A closure capturing a realm-typed variable must also be refused at
// finalize. The capture lives in a *HeapItemValue whose inner T is the
// realm pointer type, which is the saveObject signal.
var stashedFn func() string

func main(cur realm) {
	c := cur
	stashedFn = func() string {
		return c.PkgPath()
	}
	println("closure stashed, awaiting finalize panic")
}

// Error:
// cannot persist realm value: realm values are ephemeral and tied to a call frame
