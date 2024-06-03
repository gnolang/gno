package main

import (
	"gno.land/r/demo/tests"
)

func main() {
	println("IsOriginCall:", tests.IsOriginCall())
	tests.AssertOriginCall()
	println("AssertOriginCall doesn't panic when called directly")

	func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal).
		println("IsOriginCall:", tests.IsOriginCall())
		defer func() {
			r := recover()
			println("AssertOriginCall panics if when called inside a function literal:", r)
		}()
		tests.AssertOriginCall()
	}()
}

// Output:
// IsOriginCall: true
// AssertOriginCall doesn't panic when called directly
// IsOriginCall: false
// AssertOriginCall panics if when called inside a function literal: invalid non-origin call
