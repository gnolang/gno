package main

import (
	"errors"
)

func get() interface{} {
	return "hello"
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

	panic(get())
}

// Output:
// different error
