package main

import "chain/runtime"

func main() {
	defer runtime.AssertOriginCall()
}

// Error:
// invalid non-origin call
