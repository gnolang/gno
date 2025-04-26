package main

import (
	"gno.land/p/demo/testutils"
)

type mystruct struct{}

func (_ mystruct) privateMethod() string {
	return "mystruct.privateMethod"
}

type PrivateInterface2 interface {
	privateMethod() string
}

func main() {
	var x PrivateInterface2 = mystruct{}
	testutils.PrintPrivateInterface(x)
}

// Error:
// main/files/access7.gno:19:2: main.PrivateInterface2 does not implement gno.land/p/demo/testutils.PrivateInterface (missing method privateMethod)

// TypeCheckError:
// main/files/access7.gno:19:34: cannot use x (variable of interface type PrivateInterface2) as testutils.PrivateInterface value in argument to testutils.PrintPrivateInterface: PrivateInterface2 does not implement testutils.PrivateInterface (missing method privateMethod)
