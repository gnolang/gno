// PKGPATH: gno.land/r/test
package test

import "chain/runtime"

func Register() {
	runtime.AssertOriginCall()
}

func helper() {
	Register()
}

func main() {
	// Closures in the realm are transparent.
	func() {
		Register()
	}()
	println("closure call ok")

	func() {
		func() {
			Register()
		}()
	}()
	println("nested closure call ok")

	// A named function in the same realm is transparent too: nothing on the
	// stack runs with another realm's storage authority.
	helper()
	println("named func call ok")
}

// Output:
// closure call ok
// nested closure call ok
// named func call ok
