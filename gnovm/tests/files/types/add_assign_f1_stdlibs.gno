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
// 1. base type of left is int64, op is legal;
// 2. while RHS is interface kind, and can be converted to left
func main() {
	r := Error(0)
	r += errCmp // in case of this, should panic mismatch on operand, except RHS is untyped
	//println(r)
	// println(Error(0) == errCmp) // Note: this is different with +=

}

// Error:
// main/files/types/add_assign_f1_stdlibs.gno:21:2: invalid operation: mismatched types main.Error and .uverse.error
