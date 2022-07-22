package main

import (
	"gno.land/p/testutils"
)

func main() {
	x := testutils.NewTestAccessStruct("PUB", "PRIV")
	println(x.privateField)
}

// Error:
// main/files2/access4.gno:9: cannot access gno.land/p/testutils.TestAccessStruct.privateField from main