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

// specil case:
// one is interface
func main() {
	if Error(0) == errCmp {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// something else
