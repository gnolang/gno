package main

import (
	"errors"
)

var errCmp1 = errors.New("XXXX")
var errCmp2 = errors.New("XXXX")

// specil case:
// one is interface
func main() {
	if errCmp1 > errCmp2 {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/files/types/eql_0f27_stdlibs.gno:13: operator > not defined on: InterfaceKind
