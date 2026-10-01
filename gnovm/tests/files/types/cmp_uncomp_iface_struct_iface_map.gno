package main

// Nested interface boundary: outer interface holds a concrete struct whose
// field is itself interface-typed and carries a map. Go's runtime re-enters
// interface-equality at the field, so the panic names the inner map type,
// not the outer struct type.
type S struct{ a interface{} }

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var m map[int]int
	var x interface{} = S{a: m}
	_ = x == x
}

// Output:
// recover: runtime error: comparing uncomparable type map[int]int
