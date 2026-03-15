package main

func A() int { return b }
func B() int { return a }

var a = func() int { return A() }()
var b = func() int { return B() }()

func main() {
	println(a, b)
}

// Error:
// main/recursive10.gno:6:5: circular dependency: a -> A -> b -> B -> a

// TypeCheckError:
// main/recursive10.gno:6:5: initialization cycle for a; main/recursive10.gno:6:5: 	a refers to A; main/recursive10.gno:3:6: 	A refers to b; main/recursive10.gno:7:5: 	b refers to B; main/recursive10.gno:4:6: 	B refers to a
