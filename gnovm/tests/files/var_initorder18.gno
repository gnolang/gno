package main

// Tests that a self-referencing variable through its own function-literal
// initializer is detected as a circular dependency. B's init expression
// assigns to B, creating B -> B.

var B = func() int {
	B = 1
}()

func main() { println(B) }

// Error:
// main/var_initorder18.gno:7:5: circular dependency: B -> B

// TypeCheckError:
// main/var_initorder18.gno:9:1: missing return; main/var_initorder18.gno:7:5: initialization cycle: B refers to itself
