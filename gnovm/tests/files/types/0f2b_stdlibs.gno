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
	if Error(0) <= errCmp {
		//if errCmp == 1 {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/files/types/0f2b_stdlibs.gno:19: operator <= not defined on: .uverse.error
