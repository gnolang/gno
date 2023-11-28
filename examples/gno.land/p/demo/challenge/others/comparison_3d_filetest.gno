package main

import (
	"strconv"
)

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

type Error2 int64

func (e Error2) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

type Error3 int64

func (e Error3) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

func main() {
	var e1 Error1
	var e2 Error2

	e1, e2 = 0, 0

	println(e1, e2)
}

// Output:
// error: 0 error: 0
