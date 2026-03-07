package main

func main() {

}

var a = func() {
	b()
}

var b = func() {
	a()
}

// Error:
// main: loop in variable initialization: remaining declarations [a b] have circular dependencies

// TypeCheckError:
// main/closure.gno:7:5: initialization cycle for a; main/closure.gno:7:5: 	a refers to b; main/closure.gno:11:5: 	b refers to a
