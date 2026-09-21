// PKGPATH: gno.land/r/demo/curlegal
package curlegal

// POSITIVE CONTROL for the three rejections: everything here is legal and must
// stay legal. Without this, tightening any of the rules into an over-rejection
// would pass CI.

func helper(cur realm) string { return cur.PkgPath() }

func main(cur realm) {
	// threading cur unchanged into a no-cross crossing helper
	println("threaded:", helper(cur))
	// and through cross()
	println("crossed:", helper(cross(cur)))

	// the DEFINE form of range binds a NEW name; it is not a write to the
	// parameter slot and cannot be cur-called, so it stays legal.
	for _, cur := range []realm{cur.Previous()} {
		println("range-define pkgpath:", cur.PkgPath())
	}

	// a non-realm variable named `cur` is untouched by all three rules,
	// including taking its address.
	{
		cur := 41
		p := &cur
		*p = *p + 1
		for _, cur = range []int{7} {
		}
		println("int cur:", cur)
	}
}

// Output:
// threaded: gno.land/r/demo/curlegal
// crossed: gno.land/r/demo/curlegal
// range-define pkgpath:
// int cur: 7
