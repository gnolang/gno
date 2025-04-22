package main

import (
	"gno.land/p/demo/testutils"
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
// main/files/access6.gno:15:2: main.mystruct does not implement gno.land/p/demo/testutils.PrivateInterface (missing method privateMethod)

// TypeCheckError:
// main/files/access6.gno:15:34: cannot use x (variable of struct type mystruct) as testutils.PrivateInterface value in argument to testutils.PrintPrivateInterface: mystruct does not implement testutils.PrivateInterface (unexported method privateMethod)
