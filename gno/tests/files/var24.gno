package main

func main() {
	var i interface{} = 1
	var a, b, c = i.(int)
}

// Error:
// main/files/var24.gno:5:6: assignment mismatch: 3 variable(s) but 1 value(s)
