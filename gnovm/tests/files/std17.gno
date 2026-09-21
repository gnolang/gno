// PKGPATH: gno.land/r/test
package test

import "chain/runtime"

// The test's main stands in for the message, and a message cannot assert:
// a basic frame between main and the assertion does not make it a call.
func main() {
	defer func() {
		println("recovered:", recover())
	}()
	for i := 0; i < 1; i++ {
		runtime.AssertOriginCall()
	}
}

// Output:
// recovered: invalid non-origin call
