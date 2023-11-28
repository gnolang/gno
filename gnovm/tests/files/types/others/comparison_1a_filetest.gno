package main

import (
	"strconv"
)

type Error int64

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// left is const, right is const, left is untyped
func main() {
	if 1 == Error(1) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// what the firetruck?
