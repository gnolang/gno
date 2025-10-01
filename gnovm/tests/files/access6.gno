package main

import (
	"gno.land/p/nt/testutils"
)

type mystruct struct{}

func (_ mystruct) privateMethod() string {
	return "mystruct.privateMethod"
}

func main() {
	x := mystruct{}
	testutils.PrintPrivateInterface(x)
}

// Error:
// main/access6.gno:15:2-36: main.mystruct does not implement gno.land/p/nt/testutils.PrivateInterface (missing method privateMethod)

// TypeCheckError:
// main/access6.gno:15:34: cannot use x (variable of struct type mystruct) as testutils.PrivateInterface value in argument to testutils.PrintPrivateInterface: mystruct does not implement testutils.PrivateInterface (unexported method privateMethod)
