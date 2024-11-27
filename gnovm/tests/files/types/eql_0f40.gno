package main

import (
	"errors"
)

type animal interface {
	eat()
}

type dog struct {
}

func (d *dog) eat() {
	println("dog eating")
}

func get() animal {
	d := &dog{}
	return d
}

var errCmp = errors.New("errCmp")

// no empty interface, different interface(with different methods)
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
