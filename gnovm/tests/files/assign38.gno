package main

func main() {
	a := 1
	a = nil
	println(a)
}

// Error:
// main/files/assign38.gno:5:2: cannot use nil as int value in assignment

// TypeCheckError:
// main/files/assign38.gno:5:6: cannot use nil as int value in assignment
