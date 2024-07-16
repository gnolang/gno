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
	if get() == errCmp {
		println("same error")
	} else {
		println("different error")
	}
}

// Error:
// main/files/types/eql_0f41_stdlibs.gno:27:5: main.animal does not implement .uverse.error
