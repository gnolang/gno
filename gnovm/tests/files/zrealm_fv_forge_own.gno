// PKGPATH: gno.land/r/fvforgeown
package fvforgeown

type Own struct{ N int }

// conv runs non-crossing: converting a local struct to this realm's own
// declared type stays legal.
func conv() *Own {
	return (*Own)(&struct{ N int }{7})
}

func main(cur realm) {
	println(conv().N)
}

// Output:
// 7
