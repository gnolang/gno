package main

import (
	"errors"
)

func foo() (err error) {
	y := 1
	defer func() {
		if r := recover(); r != nil {
			switch y {
			case 1:
				err = errors.New("ok")
			default:
				err = nil
			}
		}
	}()
	panic(y)
}

func main() {
	println(foo())
}

// Output:
// ok
