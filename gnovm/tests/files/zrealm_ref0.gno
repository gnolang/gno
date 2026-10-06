// PKGPATH: gno.land/r/test
package test

// Realm filetest: ref of a concrete type in a realm.
// This works correctly since xv.TV.T matches the static type.

var x int

func init() {
	x = 1
}

func main(cur realm) {
	p := &x
	*p = 2
	println(x)
}

// Output:
// 2
