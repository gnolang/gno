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

	var r Robot

	r = animal.(Dog)

	println(r)
}

// Error:
// main/assign_type_assertion_d.gno:19:2-18: cannot use main.Dog as main.Robot without explicit conversion

// TypeCheckError:
// main/assign_type_assertion_d.gno:19:6: cannot use animal.(Dog) (comma, ok expression of struct type Dog) as Robot value in assignment
