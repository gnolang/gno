package main

import (
	"gno.land/p/testutils"
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
// main.mystruct does not implement gno.land/p/testutils.PrivateInterface
