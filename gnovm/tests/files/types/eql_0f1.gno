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
	if int64(1) == errCmp {
		//if errCmp == 1 {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/eql_0f1.gno:19:5-23: invalid operation: (mismatched types int64 and .uverse.error)

// TypeCheckError:
// main/eql_0f1.gno:19:17: invalid operation: int64(1) == errCmp (mismatched types int64 and error)
