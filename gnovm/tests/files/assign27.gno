package main

func main() {
	s := []string{"1", "2"}
	a, b, c := s[0]
}

// Error:
// main/assign27.gno:5:2-17: assignment mismatch: 3 variable(s) but 1 value(s)

// TypeCheckError:
// main/assign27.gno:5:13: assignment mismatch: 3 variables but 1 value; main/assign27.gno:5:2: declared and not used: a; main/assign27.gno:5:5: declared and not used: b; main/assign27.gno:5:8: declared and not used: c
