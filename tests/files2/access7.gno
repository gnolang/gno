package main

import (
	"gno.land/p/testutils"
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
// main.PrivateInterface2 does not implement gno.land/p/testutils.PrivateInterface
