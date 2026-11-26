// PKGPATH: gno.land/r/tests/vm/t55spotpkg
package t55spotpkg

var count int

// O1: a literal called where it is written, as a package-level initialiser,
// runs as the realm's own code; so does a local helper inside it.
var a = func() int { count++; return count }()
var b = func() int {
	f := func() { count++ }
	f()
	f()
	return count
}()

func main(cur realm) {
	println("s01 package-level on-the-spot literal", a)
	println("s02 local helper inside it", b, count)
}

// Output:
// s01 package-level on-the-spot literal 1
// s02 local helper inside it 3 3
