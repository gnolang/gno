package main

import (
	"gno.land/p/nt/testutils"
)

func main() {
	// Keyed struct literal with unexported field from external package.
	x := testutils.TestAccessStruct{PublicField: "PUB", privateField: "PRIV"}
	println(x)
}

// Error:
// main/access8.gno:9:54-66: cannot refer to unexported field privateField in struct literal of type struct{PublicField string; privateField string}

// TypeCheckError:
// main/access8.gno:9:54: cannot refer to unexported field privateField in struct literal of type testutils.TestAccessStruct
