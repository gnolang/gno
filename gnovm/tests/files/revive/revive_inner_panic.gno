// PKGPATH: gno.land/r/crossrealm
package crossrealm

// revive's fn panics E2, and its defer crosses into f, which panics F; the
// boundary frame f holds E2, the revive frame holds E. revive aborts both E2
// and F, so E alone must reach main's recover.
func g(cur realm) {
	f := func(cur realm) {
		panic("F")
	}
	r := revive(func() {
		defer func() { f(cross(cur)) }()
		panic("E2")
	})
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
