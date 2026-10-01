// Challenge scratch: the outer package declares its OWN unexported method sec
// (main-qualified) shadowing the promoted zzsecx-qualified sec at depth 1.
// Oracle: Go compiles this and secx.UseSec(FM{}) dispatches the promoted
// secx.sec => 7 (identity-based selection, not name-depth).
package main

import "filetests/extern/secx"

type FM struct {
	secx.SecImpl
}

func (FM) sec() int { return 1 }

func main() {
	println("A:", secx.UseSec(FM{}))
}

// Output:
// A: 7
