package main

import (
	"gno.land/p/testutils"
)

func main() {
	println(testutils.testVar2)
}

// Error:
// cannot access gno.land/p/testutils.testVar2 from main
