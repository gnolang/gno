// PKGPATH: gno.land/r/crossrealm
package crossrealm

func main(cur realm) {
	f := func(cur realm) {
		panic("panicked!")
	}
	abort(0, cur, f)
}

func abort(_ int, rlm realm, f any) {

	var didAbort bool
	var abortValue any
	var r any

	switch f := f.(type) {
	case func():
		r = revive(f)
	case func(realm):
		r = revive(func() { f(cross2(rlm)) })
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
