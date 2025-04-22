package main

func main() {
	s := []string{"1", "2"}
	var a, b, c = s[0]
}

// Error:
// main/files/var25.gno:5:6: assignment mismatch: 3 variable(s) but 1 value(s)

// TypeCheckError:
// main/files/var25.gno:5:16: assignment mismatch: 3 variables but 1 value; main/files/var25.gno:5:6: declared and not used: a; main/files/var25.gno:5:9: declared and not used: b; main/files/var25.gno:5:12: declared and not used: c
