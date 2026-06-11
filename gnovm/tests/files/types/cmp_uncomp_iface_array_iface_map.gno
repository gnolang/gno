package main

// Nested interface boundary: outer interface holds an array of interface{}
// elements, one of which carries a map. Go re-enters interface-equality at
// each element, so the panic names the inner map type — not the outer
// [N]interface{} carrier.
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var m map[int]int
	var x interface{} = [2]interface{}{1, m}
	_ = x == x
}

// Output:
// recover: runtime error: comparing uncomparable type map[int]int
