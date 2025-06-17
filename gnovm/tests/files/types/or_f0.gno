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
	println(1 | errCmp)
}

// Error:
// main/or_f0.gno:19:10-20: operator | not defined on: InterfaceKind

// TypeCheckError:
// main/or_f0.gno:19:10: cannot convert 1 (untyped int constant) to type interface{Error() string}
