package main

import (
	"errors"
)

func main() {
	println(f(false))
	println(f(true))
}

func f(dopanic bool) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = x.(error)
		}
	}()
	q(dopanic)
	return
}

func q(dopanic bool) {
	if dopanic {
		panic(errors.New("wtf"))
	}
}

// Output:
// undefined
// wtf

