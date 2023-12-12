package main

import (
	"strconv"
)

type Error int8

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// left is untyped const, right is not const

var a int8

func main() {
	a = 1
	if 1 == a {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// what the firetruck?
