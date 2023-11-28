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

func main() {
	e1 := Error1(0)
	e2 := Error2(0)
	e1 += e2
}

// Error:
// main/gno.land/p/demo/challenge/comparison_3b_filetest.gno:22: mismatch type for Assign
