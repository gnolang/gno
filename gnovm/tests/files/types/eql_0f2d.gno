package main

import (
	"fmt"
	"strconv"
)

type E interface {
	Error() string
}
type Error int64

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// special case:
// one is interface
func main() {
	var e0 E
	e0 = Error(0)
	fmt.Printf("%T\n", e0)
	if e0 == Error(0) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// int64
// what the firetruck?
