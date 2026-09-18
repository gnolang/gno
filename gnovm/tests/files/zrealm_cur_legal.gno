// PKGPATH: gno.land/r/demo/curlegal
package curlegal

// The fixed-binding rules key on the resolved type, so everything that is not
// a realm-typed binding named `cur` is left alone: threading `cur` unchanged
// into a no-cross crossing helper, crossing with cross(cur), a realm-typed
// binding under another name, and a `cur` of a non-realm type.

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

	// a non-realm variable named `cur` is untouched by every rule.
	{
		cur := 41
		p := &cur
		*p = 42
		for _, cur = range []int{7} {
		}
		println("int cur:", cur)
	}
	// a type-switch clause var named `cur` is fine when the clause type is
	// not realm.
	switch cur := any(7).(type) {
	case int:
		println("typeswitch int cur:", cur)
	}
}

// Output:
// threaded: gno.land/r/demo/curlegal
// crossed: gno.land/r/demo/curlegal
// range pkgpath: gno.land/r/demo/curlegal
// int cur: 7
// typeswitch int cur: 7
