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

// special case:
// one is interface
func main() {
	if Error(0) < errCmp {
		//if errCmp == 1 {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/eql_0f2c.gno:19:5-22: operator < not defined on: InterfaceKind

// TypeCheckError:
// main/eql_0f2c.gno:19:16: invalid operation: Error(0) < errCmp (operator < not defined on interface)
