package main

func main() {
	var i any = 1
	a, b, c := i.(int)
}

// Error:
// main/assign26.gno:5:2-20: assignment mismatch: 3 variable(s) but 1 value(s)

// TypeCheckError:
// main/assign26.gno:5:13: assignment mismatch: 3 variables but 1 value; main/assign26.gno:5:2: declared and not used: a; main/assign26.gno:5:5: declared and not used: b; main/assign26.gno:5:8: declared and not used: c
