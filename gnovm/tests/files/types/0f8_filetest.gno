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
	if errCmp == int64(1) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/files/types/0f8_filetest.gno:19: int64 does not implement .uverse.error
