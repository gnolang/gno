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
// main/eql_0f27.gno:13:5-22: operator > not defined on: InterfaceKind

// TypeCheckError:
// main/eql_0f27.gno:13:5: invalid operation: errCmp1 > errCmp2 (operator > not defined on interface)
