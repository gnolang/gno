package main

import (
	"gno.land/p/testutils"
)

func main() {
	x := testutils.NewTestAccessStruct("PUB", "PRIV")
	println(x.PublicMethod())
}

// Output:
// PUB/PRIV
