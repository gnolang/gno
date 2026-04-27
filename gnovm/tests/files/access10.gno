package main

import (
	"gno.land/p/nt/testutils/v0"
)

func main() {
	// Keyed struct literal with non-existent lowercase field from external package.
	x := testutils.TestAccessStruct{typoField: "val"}
	println(x)
}

// Error:
// main/access10.gno:9:34-43: unknown field typoField in struct literal of type gno.land/p/nt/testutils/v0.TestAccessStruct

// TypeCheckError:
// main/access10.gno:9:34: unknown field typoField in struct literal of type testutils.TestAccessStruct
