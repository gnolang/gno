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

	var ok bool

	1, ok = animal.(Dog)

	println(ok)
}

// Error:
// main/files/types/assign_type_assertion_f.gno:19: cannot use main.Dog as <untyped> bigint without explicit conversion
