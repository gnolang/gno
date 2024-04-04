package main

import (
	"strconv"
)

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// both not const, typed
func main() {
	var e1 Error1 = Error1(0)
	var e2 Error1 = Error1(0)
	println(e1 / e2)
}

// Error:
// runtime error: integer divide by zero
