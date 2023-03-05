package main

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func main() {
	test()
	std.TestSetOrigCaller(testutils.TestAddress("user"))
	test()
}

func test() {
	if std.IsOriginCall() {
		println("IsOriginCall: true")
	} else {
		println("IsOriginCall: false")
	}
	defer func() {
		if r := recover(); r != nil {
			println("AssertOriginCall: panic:", r.(string))
		}
	}()
	std.AssertOriginCall()
	println("AssertOriginCall: no panic")
}

// Output:
// IsOriginCall: false
// AssertOriginCall: panic: invalid non-origin call
// IsOriginCall: true
// AssertOriginCall: no panic
