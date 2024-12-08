package main

import (
	"strconv"
)

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

func main() {
	var e1 Error1 = Error1(0)
	var e2 int64 = 1
	e1 += Error1(e2)
	println(e1)
}

// Output:
// error: 1
