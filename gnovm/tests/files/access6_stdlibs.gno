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
// main/files/access6_stdlibs.gno:15: main.mystruct does not implement gno.land/p/demo/testutils.PrivateInterface
