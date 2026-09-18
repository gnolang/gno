// PKGPATH: gno.land/r/demo/curlegal
package curlegal

// The fixed-binding rules key on the resolved type, so everything that is not
// a realm-typed binding named `cur` is left alone: threading `cur` unchanged
// into a no-cross crossing helper, crossing with cross(cur), a realm-typed
// binding under another name.

func helper(cur realm) string { return cur.PkgPath() }

func main(cur realm) {
	// threading cur unchanged into a no-cross crossing helper
	println("threaded:", helper(cur))

	println("crossed:", helper(cross(cur)))

	// a realm-typed binding under another name is an ordinary realm value,
	// not any frame's identity; declaring it is fine.
	for _, prev := range []realm{cur} {
		println("range pkgpath:", prev.PkgPath())
	}

}

// Output:
// threaded: gno.land/r/demo/curlegal
// crossed: gno.land/r/demo/curlegal
// range pkgpath: gno.land/r/demo/curlegal
