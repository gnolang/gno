package main

import (
	"strconv"
)

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

type Error2 int64

func (e Error2) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

func main() {
	var e1 Error1
	e1 = Error2(0)
}

// Error:
// main/gno.land/p/demo/challenge/comparison_3c_filetest.gno:21: mismatch type for Assign, in case: a, b = x, y
