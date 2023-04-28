package main

import (
	"gno.land/p/demo/testutils"
)

func main() {
	x := testutils.NewTestAccessStruct("PUB", "PRIV")
	println(x.privateField)
}

// Error:
// main/files/access4.gno:9: cannot access gno.land/p/demo/testutils.TestAccessStruct.privateField from main
