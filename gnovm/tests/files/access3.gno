package main

import (
	"gno.land/p/nt/testutils"
)

func main() {
	x := testutils.NewTestAccessStruct("PUB", "PRIV")
	println(x.PublicField)
}

// Output:
// PUB
