// PKGPATH: gno.land/r/crossrealm
package crossrealm

// revive catches F inside g, called by a defer while E is live; a runtime
// fault in g then links onto E and must find no stale link left from F.
func g(cur realm) {
	f := func(cur realm) {
		panic("F")
	}
	r := revive(func() { f(cross(cur)) })
	println("revived:", r)
	var a []int
	i := 1
	_ = a[i]
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
// main recovered: runtime error: nil slice index (out of bounds)
