package main

import (
	"strconv"
)

type E interface {
	Error() string
}

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

type Error2 int

func (e Error2) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// both not const
func main() {
	var e1 E = Error1(0)
	var e2 E = Error2(0)
	if e1 == e2 {
		println("what the firetruck?")
	} else {
		// supposed print this
		println("something else")
	}
}

// Output:
// something else
