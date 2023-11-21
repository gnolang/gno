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
	var e2 Error2
	var b int32 = 0

	e1, e2 = 0, b

	println(e1, e2)
}

// Error:
// main/gno.land/p/demo/challenge/comparison_3g_filetest.gno:24: mismatch type for Assign, in case: a, b = x, y
