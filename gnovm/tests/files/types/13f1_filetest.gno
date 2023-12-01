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
	r += errCmp
	println(r)
}

// Error:
// invalid operation: mismatched types main.Error and *errors.errorString
