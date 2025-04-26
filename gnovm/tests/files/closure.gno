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
// main/files/closure.gno:7:5: constant definition loop with a

// TypeCheckError:
// main/files/closure.gno:7:5: initialization cycle for a; main/files/closure.gno:7:5: 	a refers to b; main/files/closure.gno:11:5: 	b refers to a
