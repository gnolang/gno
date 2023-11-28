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
	println(1 + errCmp)
}

// Error:
// main/files/types/1f0_filetest.gno:19: + should have operand number or string, while have: .uverse.error
