package main

import (
	"gno.land/r/demo/tests"
)

func main() {
	println("tests.CallIsOriginCall:", cross(tests.CallIsOriginCall)())
	cross(tests.CallAssertOriginCall)()
	println("tests.CallAssertOriginCall doesn't panic when called directly")

	{
		// if called inside a block, this is no longer an origin call because
		// there's one additional frame (the block).
		println("tests.CallIsOriginCall:", cross(tests.CallIsOriginCall)())
		defer func() {
			r := recover()
			println("tests.AssertOriginCall panics if when called inside a function literal:", r)
		}()
		cross(tests.CallAssertOriginCall)()
	}
}

// Output:
// tests.CallIsOriginCall: false
// tests.CallAssertOriginCall doesn't panic when called directly
// tests.CallIsOriginCall: false
// tests.AssertOriginCall panics if when called inside a function literal: undefined
