package main

import (
	"strconv"
)

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

type Error2 int

func (e Error2) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// both typed(different) const
func main() {
	println(Error1(0) == Error2(0))
}

// Error:
// main/files/types/eql_0a2.gno:21:10: cannot use main.Error1 as main.Error2 without explicit conversion
