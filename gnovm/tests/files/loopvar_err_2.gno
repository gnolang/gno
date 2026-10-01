package main

// Using a loop variable after the loop exits should report "name not defined: i"
// not "name not defined: i.loopvar".
func main() {
	for i := 0; i < 3; i++ {
		_ = i
	}
	println(i) // i is out of scope here
}

// Error:
// main/loopvar_err_2.gno:9:10-11: name i not declared

// TypeCheckError:
// main/loopvar_err_2.gno:9:10: undefined: i
