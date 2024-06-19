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
// main/files/types/add_assign_f0_stdlibs.gno:20: invalid operation: mismatched types int and .uverse.error
