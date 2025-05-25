package main

func main() {
	var i any = 1
	var a, b, c = i.(int)
}

// Error:
// main/var24.gno:5:6-23: assignment mismatch: 3 variable(s) but 1 value(s)

// TypeCheckError:
// main/var24.gno:5:16: assignment mismatch: 3 variables but 1 value; main/var24.gno:5:6: declared and not used: a; main/var24.gno:5:9: declared and not used: b; main/var24.gno:5:12: declared and not used: c
