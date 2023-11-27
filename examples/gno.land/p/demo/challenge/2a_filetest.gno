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

// left is typed const, right is not const, should not compile
func main() {
	if int(1) == errCmp {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/gno.land/p/demo/challenge/2a_filetest.gno:18: int does not implement .uverse.error
