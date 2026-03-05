package main

import (
	tests "gno.land/r/tests/vm"
)

func main() {
	println("tests.CallIsOriginCall:", tests.CallIsOriginCall(cross))
	tests.CallAssertOriginCall(cross)
	println("tests.CallAssertOriginCall doesn't panic when called directly")

	{
		// if called inside a block, this is no longer an origin call because
		// there's one additional frame (the block).
		println("tests.CallIsOriginCall:", tests.CallIsOriginCall(cross))
		defer func() {
			r := recover()
			println("tests.AssertOriginCall panics if when called inside a function literal:", r)
		}()
		tests.CallAssertOriginCall(cross)
	}
}

// Output:
// tests.CallIsOriginCall: false
// tests.CallAssertOriginCall doesn't panic when called directly
// tests.CallIsOriginCall: false
// tests.AssertOriginCall panics if when called inside a function literal: undefined
