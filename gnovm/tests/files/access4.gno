package main

import (
	"gno.land/p/demo/testutils"
)

func main() {
	x := testutils.NewTestAccessStruct("PUB", "PRIV")
	println(x.privateField)
}

// Error:
// main/access4.gno:9:10-24: cannot access gno.land/p/demo/testutils.TestAccessStruct.privateField from main

// TypeCheckError:
// main/access4.gno:9:12: x.privateField undefined (cannot refer to unexported field privateField)
