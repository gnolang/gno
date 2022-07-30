package main

import (
	"gno.land/p/testutils"
)

func main() {
	println(testutils.testVar2)
}

// Error:
// main/files2/access1.gno:8: cannot access gno.land/p/testutils.testVar2 from main