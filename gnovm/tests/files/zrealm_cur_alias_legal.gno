// PKGPATH: gno.land/r/demo/curaliasok
package curaliasok

// Positive control: an alias for realm is not banned, it is just held to the
// same naming rules as the spelling. A crossing function declared through one
// works normally.

type R = realm

func inner(cur R) string { return cur.PkgPath() }

func main(cur realm) {
	println("aliased:", inner(cur))
	println("aliased crossed:", inner(cross(cur)))
}

// Output:
// aliased: gno.land/r/demo/curaliasok
// aliased crossed: gno.land/r/demo/curaliasok
