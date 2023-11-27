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

// both not const, should not compile
func main() {
	var a int = 1
	if a == errCmp {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/gno.land/p/demo/challenge/2b_filetest.gno:19: int does not implement .uverse.error
