package main

// Type error inside a for loop: check that the error message does NOT
// expose the internal ".loopvar" name suffix to users.
func main() {
	for i := 0; i < 3; i++ {
		var s string = i
		_ = s
	}
}

// Error:
// main/loopvar_err_1.gno:7:7-19: cannot use int as string

// TypeCheckError:
// main/loopvar_err_1.gno:7:18: cannot use i (variable of type int) as string value in variable declaration
