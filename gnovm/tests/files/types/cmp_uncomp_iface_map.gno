package main

// Nil-wrapped map via interface{} must panic, matching Go's runtime
// semantics. Also verifies the panic message format.
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var m map[int]int
	var x interface{} = m
	_ = x == x
}

// Output:
// recover: runtime error: comparing uncomparable type map[int]int
