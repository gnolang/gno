package main

import (
	"strconv"
)

type Error int8

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// left is typed const, right is untyped const
func main() {
	if Error(1) == 1 {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// what the firetruck?
