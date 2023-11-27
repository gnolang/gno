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

// left is untyped const, right is not const
// compile fail since the untyped 1 doesn't conform error interface
func main() {
	if 1 == errCmp {
		//if errCmp == 1 {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/gno.land/p/demo/challenge/2_filetest.gno:19: <untyped> bigint does not implement .uverse.error
