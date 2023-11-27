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

func main() {
	if errCmp == int(1) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/gno.land/p/demo/challenge/2c_filetest.gno:17: int does not implement .uverse.error
