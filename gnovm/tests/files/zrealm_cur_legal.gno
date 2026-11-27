// PKGPATH: gno.land/r/demo/curlegal
package curlegal

// `cur` is reserved by name, but only its declaration is: threading `cur`
// unchanged into a no-cross crossing helper, crossing with cross(cur), and a
// realm-typed binding under another name are all ordinary uses.

func helper(cur realm) string { return cur.PkgPath() }

// A function type or interface method may name its first realm parameter
// `cur`: it is the crossing shape, and the check carves it out by position.
type Crosser interface{ Do(cur realm) string }

type CrossFn func(cur realm) string

type impl struct{}

func (impl) Do(cur realm) string { return cur.PkgPath() }

func main(cur realm) {
	// threading cur unchanged into a no-cross crossing helper
	println("threaded:", helper(cur))

	println("crossed:", helper(cross(cur)))

	var c Crosser = impl{}
	var f CrossFn = helper
	println("iface:", c.Do(cur), "fn:", f(cur))

	// a realm-typed binding under another name is an ordinary realm value,
	// not any frame's identity; declaring it is fine.
	for _, prev := range []realm{cur} {
		println("range pkgpath:", prev.PkgPath())
	}

}

// Output:
// threaded: gno.land/r/demo/curlegal
// crossed: gno.land/r/demo/curlegal
// iface: gno.land/r/demo/curlegal fn: gno.land/r/demo/curlegal
// range pkgpath: gno.land/r/demo/curlegal
