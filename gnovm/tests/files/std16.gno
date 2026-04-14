package main

import "chain/runtime"

func Register() {
	runtime.AssertOriginCall()
}

func callFunc(f func()) {
	f()
}

func helper() {
	Register()
}

func main() {
	// Security case 1: closure passed as callback to a named function.
	// main → callFunc → Register → AssertOriginCall (4 call frames > 3).
	// The closure is transparent, but callFunc is a real call boundary — SHOULD panic.
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		callFunc(func() {
			Register()
		})
	}()
	println("callback to named func panicked:", panicked)

	// Security case 2: named function wrapped in closure.
	// main → helper → Register → AssertOriginCall (4 call frames > 3).
	// The outer closure is transparent, but helper is a real call — SHOULD panic.
	panicked = false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		helper()
	}()
	println("named func in closure panicked:", panicked)

	// Sanity: direct closure call — should NOT panic.
	// main → Register → AssertOriginCall (3 call frames).
	func() {
		Register()
	}()
	println("direct closure ok")
}

// Output:
// callback to named func panicked: true
// named func in closure panicked: true
// direct closure ok
