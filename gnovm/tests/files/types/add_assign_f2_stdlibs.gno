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
	r := Error(0)
	errCmp += r
	println(errCmp)
}

// Error:
// main/files/types/add_assign_f2_stdlibs.gno:20: operator += not defined on: InterfaceKind
