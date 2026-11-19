// PKGPATH: gno.land/r/crossrealm
package crossrealm

// A deferred revive catches F while E is live; E must still reach main's recover.
func main(cur realm) {
	f := func(cur realm) {
		panic("F")
	}
	defer func() {
		println("main recovered:", recover())
	}()
	defer revive(func() { f(cross(cur)) })
	panic("E")
}

// Output:
// main recovered: E
