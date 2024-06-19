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
// main/files/types/assign_type_assertion_g.gno:20: want bool type got int
