// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

func main() {
	a := func() { // in this case, not attached.
		println("func lit")
	}
	println("ok")
}

// Output:
// ok
