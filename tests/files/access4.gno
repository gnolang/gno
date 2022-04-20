package main

import (
	"gno.land/p/testutils"
)

func main() {
	x := testutils.NewTestAccessStruct("PUB", "PRIV")
	println(x.privateField)
}

// Error:
// cannot access gno.land/p/testutils.TestAccessStruct.privateField from main
