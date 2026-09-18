package main

// Concrete uncomparable struct (concrete map field, not interface-typed)
// compared via interface, where an earlier comparable field differs.
// Go checks the dynamic type's comparability first and panics naming the
// struct — it does NOT short-circuit on the differing field.
type Q struct {
	a int
	m map[int]int
}

func main() {
	defer func() { println("recover:", recover()) }()
	var x, y interface{} = Q{1, nil}, Q{2, nil}
	_ = x == y
}

// Output:
// recover: runtime error: comparing uncomparable type main.Q
