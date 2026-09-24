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
	// Closures in the realm are transparent, at any depth.
	func() {
		runtime.AssertOriginCall()
	}()
	println("closure call ok")

	func() {
		func() {
			runtime.AssertOriginCall()
		}()
	}()
	println("nested closure call ok")

	// helper is the entry; Register is a named function of this realm above
	// it, a value a script could hand back, so it is a boundary.
	defer func() {
		println("named func call panicked:", recover())
	}()
	helper()
}

// Output:
// closure call ok
// nested closure call ok
// named func call panicked: invalid non-origin call
