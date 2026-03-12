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
// circular dependency: a -> b

// TypeCheckError:
// main/closure.gno:7:5: initialization cycle for a; main/closure.gno:7:5: 	a refers to b; main/closure.gno:11:5: 	b refers to a
