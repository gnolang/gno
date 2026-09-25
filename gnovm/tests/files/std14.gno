// PKGPATH: gno.land/r/test
package test

import "chain/runtime"

func Register() {
	runtime.AssertOriginCall()
}

func helper() {
	Register()
}

func panicked(f func()) (r any) {
	defer func() { r = recover() }()
	f()
	return nil
}

// main is the message's script, so the first function of this realm after it
// is the entry, and any function of this realm above the entry, named or
// literal, is a boundary. Same realm throughout, so recover.
func main() {
	// The closure is the entry and asserts itself.
	func() {
		runtime.AssertOriginCall()
	}()
	println("closure call ok")

	// The inner closure is this realm's function above the entry.
	println("nested closure panicked:", panicked(func() {
		func() {
			runtime.AssertOriginCall()
		}()
	}))

	// helper is the entry; Register is above it.
	println("named func call panicked:", panicked(helper))
}

// Output:
// closure call ok
// nested closure panicked: invalid non-origin call
// named func call panicked: invalid non-origin call
