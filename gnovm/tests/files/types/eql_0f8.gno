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
	if errCmp == int64(1) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/eql_0f8.gno:19:5-23: invalid operation: (mismatched types int64 and .uverse.error)

// TypeCheckError:
// main/eql_0f8.gno:19:15: invalid operation: errCmp == int64(1) (mismatched types error and int64)
