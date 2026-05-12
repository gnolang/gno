package main

import "chain/runtime"

func Register() {
	runtime.AssertOriginCall()
}

func main() {
	fn := func() {
		Register()
	}
	fn()
	println("ok")
}

// Output:
// ok
