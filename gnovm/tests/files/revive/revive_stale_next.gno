// PKGPATH: gno.land/r/crossrealm
package crossrealm

// revive catches F inside a function a defer calls while E is live; the next
// panic, P, links onto E and must find no stale link left from F.
func g(cur realm) {
	f := func(cur realm) {
		panic("F")
	}
	r := revive(func() { f(cross(cur)) })
	println("revived:", r)
}

func main(cur realm) {
	defer func() {
		r := recover()
		println("main recovered:", r)
	}()
	defer func() {
		panic("P")
	}()
	defer func() {
		g(cur)
	}()
	panic("E")
}

// Output:
// revived: F
// main recovered: P
