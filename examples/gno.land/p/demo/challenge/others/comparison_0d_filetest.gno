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

// left is const, right is const, compile fail
func main() {
	println(Error1(0) - Error2(1))
}

// Error:
// main/gno.land/p/demo/challenge/comparison_0d_filetest.gno:21: cannot use main.Error1 as main.Error2 without explicit conversion
