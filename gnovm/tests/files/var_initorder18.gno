package main

var B = func() int {
	B = 1
}()

func main() { println(B) }

// Error:
// main/var_initorder18.gno:3:5: circular dependency: B -> B

// TypeCheckError:
// main/var_initorder18.gno:5:1: missing return; main/var_initorder18.gno:3:5: initialization cycle: B refers to itself
