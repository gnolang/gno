package main

import (
	"gno.land/p/nt/testutils/v0"
)

func main() {
	x := testutils.NewTestAccessStruct("PUB", "PRIV")
	println(x.PublicField)
}

// Output:
// PUB
