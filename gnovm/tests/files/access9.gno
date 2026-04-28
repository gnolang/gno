package main

import (
	"gno.land/p/nt/testutils/v0"
)

func main() {
	// Unkeyed struct literal with unexported fields from external package.
	x := testutils.TestAccessStruct{"PUB", "PRIV"}
	println(x)
}

// Error:
// main/access9.gno:9:7-48: implicit assignment to unexported field in struct literal of type gno.land/p/nt/testutils/v0.TestAccessStruct

// TypeCheckError:
// main/access9.gno:9:41: implicit assignment to unexported field privateField in struct literal of type testutils.TestAccessStruct
