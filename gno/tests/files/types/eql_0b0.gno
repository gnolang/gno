package main

import (
	"strconv"
)

type Error int8

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// left is untyped const, right is typed const
// left is assignable to right
func main() {
	if 1 == Error(1) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// what the firetruck?
