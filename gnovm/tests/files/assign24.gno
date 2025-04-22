package main

func main() {
	a, b := 1
}

// Error:
// main/files/assign24.gno:4:2: assignment mismatch: 2 variable(s) but 1 value(s)

// TypeCheckError:
// main/files/assign24.gno:4:10: assignment mismatch: 2 variables but 1 value; main/files/assign24.gno:4:2: declared and not used: a; main/files/assign24.gno:4:5: declared and not used: b
