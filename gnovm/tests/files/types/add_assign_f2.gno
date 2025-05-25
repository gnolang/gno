package main

import (
	"errors"
	"strconv"
)

type Error int64

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

var errCmp = errors.New("XXXX")

// specil case:
// one is interface
func main() {
	r := Error(0)
	errCmp += r
	println(errCmp)
}

// Error:
// main/add_assign_f2.gno:20:2-13: operator += not defined on: InterfaceKind

// TypeCheckError:
// main/add_assign_f2.gno:20:2: invalid operation: errCmp += r (mismatched types error and Error)
