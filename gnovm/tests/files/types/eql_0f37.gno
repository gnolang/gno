package main

import (
	"fmt"
	"strconv"
)

type Error0 int64

func (e Error0) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

func get() any {
	return Error1(0)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			if r == get() {
				fmt.Println("recover Error1")
			} else {
				fmt.Println("recover Error0")
			}
		}
	}()

	panic(Error1(0))
}

// Output:
// recover Error1
