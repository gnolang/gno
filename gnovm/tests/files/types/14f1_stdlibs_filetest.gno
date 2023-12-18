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
	r := Error(0)
	r -= errCmp // in case of this, should panic mismatch on operand, except RHS is untyped
}

// Error:
// main/files/types/14f1_stdlibs_filetest.gno:18: invalid operation: mismatched types main.Error and .uverse.error
