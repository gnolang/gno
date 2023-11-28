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

// both not const, check +
func main() {
	var e1 E = Error1(0)
	println(e1 * 0)
}

// Error:
// main/gno.land/p/demo/challenge/2g_filetest.gno:20: * should have operand number, while have: main.E
