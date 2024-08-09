package main

import (
	"errors"
	"strconv"
)

type Error int64

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// typed
var errCmp error = errors.New("XXXX")

// special case:
// one is interface
func main() {
	const e Error = Error(0) // typed const
	if e == errCmp {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// something else
