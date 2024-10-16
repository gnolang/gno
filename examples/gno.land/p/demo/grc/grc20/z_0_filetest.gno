package main

import (
	banker "gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/testutils"
)

const uint64_max = 1<<64 - 1

func main() {
	b := banker.NewBanker("Dummy", "DUMMY", 6)
	var owner = testutils.TestAddress("owner")
	// Mint uint64_max - 1000 tokens for owner
	b.Mint(owner, uint64_max-1000)
	// Try to mint 1001 tokens for owner, should fail because it exceeds the limit
	println(b.Mint(owner, 1001).Error())

}

// Output:
// you can't mint more than 1000 tokens
