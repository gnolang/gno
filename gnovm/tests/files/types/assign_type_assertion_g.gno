package main

type Animal interface {
	eat()
}

type Robot struct {
}

type Dog struct{}

func (Dog) eat() {}

func main() {
	var animal Animal = Dog{}

	var a Animal
	var ok int

	a, ok = animal.(Dog)

	println(a)
	println(ok)
}

// Error:
// main/assign_type_assertion_g.gno:20:2-22: want bool type got int

// TypeCheckError:
// main/assign_type_assertion_g.gno:20:10: cannot use animal.(Dog) (untyped bool value) as int value in assignment
