// PKGPATH: gno.land/r/crossrealm
package crossrealm

// revive catches F inside g, called by a defer while E is live; E must still
// reach main's recover, as in Go.
func g(cur realm) {
	f := func(cur realm) {
		panic("F")
	}
	r := revive(func() { f(cross(cur)) })
	println("revived:", r)
}

func main(cur realm) {
	defer func() {
		println("main recovered:", recover())
	}()
	defer func() {
		g(cur)
	}()
	panic("E")
}

// Output:
// revived: F
// main recovered: E
