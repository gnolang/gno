package main

import (
	"errors"
	"strconv"
)

type Error int64

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

var errCmp = errors.New("XXXX")

// left is typed const, right is not const
// compile pass since both conform to same interface
func main() {
	if Error(0) == errCmp { // both conform same interface, compile pass, runtime evaluation not equal
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// something else
