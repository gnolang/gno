package main

func main() {
	s := []string{"1", "2"}
	var a, b, c = s[0]
}

// Error:
// main/files/var25.gno:5:6: assignment mismatch: 3 variable(s) but 1 value(s)
