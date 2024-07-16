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
// main/files/types/eql_0f8_stdlibs.gno:19:5: int64 does not implement .uverse.error
