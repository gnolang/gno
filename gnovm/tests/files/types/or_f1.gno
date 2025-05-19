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
	println(Error(0) | errCmp)
}

// Error:
// main/or_f1.gno:19:10-27: operator | not defined on: InterfaceKind

// TypeCheckError:
// main/or_f1.gno:19:10: invalid operation: Error(0) | errCmp (mismatched types Error and error)
