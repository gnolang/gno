package main

import "chain/runtime"

func Register() {
	runtime.AssertOriginCall()
}

func main() {
	// for-loop basic frame should be transparent.
	for i := 0; i < 1; i++ {
		Register()
	}
	println("for ok")

	// range basic frame should be transparent.
	for range []int{1} {
		Register()
	}
	println("range ok")

	// switch basic frame should be transparent.
	switch 1 {
	case 1:
		Register()
	}
	println("switch ok")
}

// Output:
// for ok
// range ok
// switch ok
