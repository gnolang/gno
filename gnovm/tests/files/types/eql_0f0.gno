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
	if 1 == errCmp {
		//if errCmp == 1 {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/eql_0f0.gno:19:5-16: invalid operation: (mismatched types <untyped> bigint and .uverse.error)

// TypeCheckError:
// main/eql_0f0.gno:19:5: cannot convert 1 (untyped int constant) to type interface{Error() string}
