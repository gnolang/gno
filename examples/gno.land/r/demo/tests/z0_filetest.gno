package main

import (
	"gno.land/r/demo/tests"
)

func main() {
	println("IsOriginCall:", tests.IsOriginCall())
	tests.AssertOriginCall()
	println("AssertOriginCall doesn't panic when called directly")

	println("IsOriginCall:", tests.RSubtestsIsOriginCall())
	defer func() {
		r := recover()
		println("AssertOriginCall panics if when called inside a function literal:", r)
	}()
	tests.RSubtestsAssertOriginCall()
}

// Output:
// IsOriginCall: true
// AssertOriginCall doesn't panic when called directly
// IsOriginCall: false
// AssertOriginCall panics if when called inside a function literal: invalid non-origin call
