package main

// Tests that a circular dependency through a function-literal variable
// (ValueDecl, not FuncDecl) is correctly detected. Unlike a FuncDecl,
// a var holding a func literal participates in initialization ordering,
// so f -> B -> f is a real cycle.

var f = func() int { return B }
var B = f()

func main() {}

// Error:
// main/var_initorder22.gno:8:5: circular dependency: f -> B -> f

// TypeCheckError:
// main/var_initorder22.gno:8:5: initialization cycle for f; main/var_initorder22.gno:8:5: 	f refers to B; main/var_initorder22.gno:9:5: 	B refers to f
