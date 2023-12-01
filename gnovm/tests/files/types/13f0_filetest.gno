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
	r := 1
	r += errCmp
	println(r)
}

// Error:
// main/files/types/13f0_filetest.gno:20: int does not implement .uverse.error
