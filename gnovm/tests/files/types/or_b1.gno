package main

import (
	"strconv"
)

type Error int8

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// one untyped const, one typed const
func main() {
	println(1 | Error(1))
	println(Error(1) | 1)
}

// Output:
// error: 1
// error: 1
