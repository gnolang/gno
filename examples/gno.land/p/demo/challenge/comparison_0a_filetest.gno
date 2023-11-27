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

// left is const, right is const, both typed, compile fail
func main() {
	if Error1(0) == Error2(0) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/gno.land/p/demo/challenge/comparison_0a_filetest.gno:21: cannot use main.Error1 as main.Error2 without explicit conversion
