// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

var root interface{}

func main() {
	root = func() { // in this, case, block(main) is attached in the tree
		println("func lit")
	}
	println("ok")
}

// Output:
// ok
