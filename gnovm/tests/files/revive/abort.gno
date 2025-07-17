// PKGPATH: gno.land/r/crossrealm
package crossrealm

func main() {
	f := func(cur realm) {
		panic("panicked!")
	}
	abort(f)
}

func abort(f any) {

	var didAbort bool
	var abortValue any
	var r any

	switch f := f.(type) {
	case func():
		r = revive(f)
	case func(realm):
		r = revive(func() { f(cross) })
	default:
		panic("f must be of type func() or func(realm)")
	}
	if r != nil {
		didAbort = true
		abortValue = r
	}

	if !didAbort {
		// If the function didn't abort as expected
		println("should abort, but not")
	} else {
		println("Abort with message: ")
		println(abortValue)
	}
	return
}

// Output:
// Abort with message:
// panicked!
