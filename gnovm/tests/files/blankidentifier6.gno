package main

type Animal interface {
	Sound() string
}

type Dog struct {
	name string
}

func (d Dog) Sound() string {
	return "Woof!"
}

func main() {
	var a Animal = Dog{name: "Rex"}

	v := a.(_)

	println(v)
}

// Error:
// main/blankidentifier6.gno:18:10-11: cannot use _ as value or type

// TypeCheckError:
// main/blankidentifier6.gno:18:10: cannot use _ as value or type
