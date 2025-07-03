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
// main/add_assign_f0.gno:20:2-13: invalid operation: mismatched types int and .uverse.error

// TypeCheckError:
// main/add_assign_f0.gno:20:2: invalid operation: r += errCmp (mismatched types int and error)
