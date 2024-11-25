package main

func main() {
	var i interface{} = 1
	a, b, c := i.(int)
}

// Error:
// main/files/assign26.gno:5:2: assignment mismatch: 3 variable(s) but 1 value(s)
