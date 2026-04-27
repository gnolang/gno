package main

import "chain/runtime"

func Register() {
	runtime.AssertOriginCall()
}

func helper() {
	Register()
}

func main() {
	// Case 1: closure wrapping Register — should NOT panic.
	// Closures are transparent: main → Register → AssertOriginCall (3 call frames).
	func() {
		Register()
	}()
	println("closure call ok")

	// Case 2: nested closures wrapping Register — should NOT panic.
	// All closures are transparent: main → Register → AssertOriginCall (3 call frames).
	func() {
		func() {
			Register()
		}()
	}()
	println("nested closure call ok")

	// Case 3: named function calling Register — SHOULD panic.
	// main → helper → Register → AssertOriginCall (4 call frames > 3).
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		helper()
	}()
	println("named func call panicked:", panicked)
}

// Output:
// closure call ok
// nested closure call ok
// named func call panicked: true
