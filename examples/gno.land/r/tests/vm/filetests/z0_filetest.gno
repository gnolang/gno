// PKGPATH: gno.land/r/tests/vm/filetests/z0_filetest

package z0_filetest

import (
	tests "gno.land/r/tests/vm"
)

func main(cur realm) {
	println("tests.CallIsOriginCall:", tests.CallIsOriginCall(cross2(cur)))
	tests.CallAssertOriginCall(cross2(cur))
	println("tests.CallAssertOriginCall doesn't panic when called directly")

	{
		// if called inside a block, this is no longer an origin call because
		// there's one additional frame (the block).
		println("tests.CallIsOriginCall:", tests.CallIsOriginCall(cross2(cur)))
		defer func() {
			r := recover()
			println("tests.AssertOriginCall panics if when called inside a function literal:", r)
		}()
		tests.CallAssertOriginCall(cross2(cur))
	}
}

// Output:
// tests.CallIsOriginCall: false
// tests.CallAssertOriginCall doesn't panic when called directly
// tests.CallIsOriginCall: false
// tests.AssertOriginCall panics if when called inside a function literal: undefined
