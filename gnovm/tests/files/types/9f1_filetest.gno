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
	println(Error(0) &^ errCmp)
}

// Error:
// main/files/types/9f1_filetest.gno:19: operator &^ not defined on: .uverse.error
