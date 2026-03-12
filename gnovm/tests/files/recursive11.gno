package main

var A = func() { B() }
var B = func() { A() }

func main() {
}

// Error:
// main/recursive11.gno:3:5: circular dependency: A -> B -> A

// TypeCheckError:
// main/recursive11.gno:3:5: initialization cycle for A; main/recursive11.gno:3:5: 	A refers to B; main/recursive11.gno:4:5: 	B refers to A
