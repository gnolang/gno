package main

// Zero-length array of an uncomparable element type wrapped in interface{}.
// The recursion never visits an element (length is 0), but Go still panics
// on the array type itself — the comparability of the dynamic type is a
// type-level property, not a value-level one.
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var x interface{} = [0]map[int]int{}
	_ = x == x
}

// Output:
// recover: runtime error: comparing uncomparable type [0]map[int]int
