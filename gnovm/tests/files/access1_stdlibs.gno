package main

import (
	"gno.land/p/demo/testutils"
)

func main() {
	println(testutils.testVar2)
}

// Error:
// main/files/access1_stdlibs.gno:8: cannot access gno.land/p/demo/testutils.testVar2 from main
