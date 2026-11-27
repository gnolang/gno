package main

import "gno.land/r/tests/vm/lendpickvictim"

// No realm context (as tooling): a non-crossing realm function called directly.
func main() {
	println("before:", lendpickvictim.Total())
	lendpickvictim.Plain(lendpickvictim.Get().Bump)
	println("after:", lendpickvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendpickvictim's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
