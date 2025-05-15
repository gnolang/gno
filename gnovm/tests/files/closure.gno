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
// main/closure.gno:7:5: loop in variable initialization: dependency trail [b a] circularly depends on b

// TypeCheckError:
// main/closure.gno:7:5: initialization cycle for a; main/closure.gno:7:5: 	a refers to b; main/closure.gno:11:5: 	b refers to a
