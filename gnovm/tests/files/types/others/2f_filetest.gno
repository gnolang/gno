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

// both not const, check +
func main() {
	var e1 E = Error1(0)
	var e2 E = Error2(0)
	println(e1 * e2)
}

// Error:
// main/gno.land/p/demo/challenge/2f_filetest.gno:27: * should have operand number, while have: main.E
