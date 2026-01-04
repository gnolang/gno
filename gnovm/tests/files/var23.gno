package main

var a, b, c = 1, 2

func main() {
}

// Error:
// main/var23.gno:3:5-19: assignment mismatch: 3 variable(s) but 2 value(s)

// TypeCheckError:
// main/var23.gno:3:11: missing init expr for c
