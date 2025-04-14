package main

import (
	"strconv"
)

type Error int64

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// both not const, and both interface
func main() {
	var l any
	if l == Error(0) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// something else
