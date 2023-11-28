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
	println(Error(0) * errCmp)
}

// Error:
// main/files/types/3f1_filetest.gno:19: * should have operand number, while have: .uverse.error
