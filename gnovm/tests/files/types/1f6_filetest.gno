package main

import (
	"errors"
	"strconv"
)

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

var errCmp = errors.New("errCmp")

func main() {
	defer func() {
		if r := recover(); r != nil {
			if r == errCmp {
				println("same error")
			} else {
				println("different error")
			}
		}
	}()

	panic(Error1(0))
}

// Output:
// different error
