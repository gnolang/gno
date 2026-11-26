package main

import "gno.land/r/tests/vm/chalcvictim"

// No realm context (as tooling): a non-crossing realm function called directly.
func main() {
	println("before:", chalcvictim.Total())
	chalcvictim.Plain(chalcvictim.Get().Bump)
	println("after:", chalcvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalcvictim's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
