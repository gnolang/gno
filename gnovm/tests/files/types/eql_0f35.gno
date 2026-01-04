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

func main() {
	defer func() {
		if r := recover(); r != nil {
			if r == Error1(0) {
				fmt.Println("Recovered. Error:\n", r)
			}
		}
	}()

	panic(Error1(0))
}

// Output:
// Recovered. Error:
//  error: 0
