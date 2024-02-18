package main

import (
	"gno.land/r/demo/tests"
)

func main() {
	println("tests.CallIsOriginCall:", tests.CallIsOriginCall())
	tests.CallAssertOriginCall()
	println("tests.CallAssertOriginCall doesn't panic when called directly")

	{
		// if called inside a block, this is no longer an origin call because
		// there's one additional frame (the block).
		println("tests.CallIsOriginCall:", tests.CallIsOriginCall())
		defer func() {
			r := recover()
			println("tests.AssertOriginCall panics if when called inside a function literal:", r)
		}()
		tests.CallAssertOriginCall()
	}
}

// Output:
// tests.CallIsOriginCall: true
// tests.CallAssertOriginCall doesn't panic when called directly
// tests.CallIsOriginCall: true
// tests.AssertOriginCall panics if when called inside a function literal: undefined
