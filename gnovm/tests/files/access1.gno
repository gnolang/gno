package main

import (
	"gno.land/p/nt/testutils/v0"
)

func main() {
	println(testutils.testVar2)
}

// Error:
// main/access1.gno:8:10-28: cannot access gno.land/p/nt/testutils/v0.testVar2 from main

// TypeCheckError:
// main/access1.gno:8:20: name testVar2 not exported by package testutils
