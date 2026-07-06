package main

// Array of maps via interface{}: panic must name the outer array
// type, not the inner map element type.
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	m := map[interface{}]int{}
	m[[1]map[int]int{}] = 1
}

// Output:
// recover: runtime error: hash of unhashable type [1]map[int]int
