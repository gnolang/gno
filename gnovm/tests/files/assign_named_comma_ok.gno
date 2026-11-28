package main

// A comma-ok result is an untyped bool, so it is assignable to a declared
// bool. The variable declared for it is a bool, as in Go.
// See https://github.com/gnolang/gno/issues/6493.

type B bool

func main() {
	m := map[string]B{"k": true}
	var v, ok B = m["k"]
	println(v, ok)

	v2, ok2 := m["k"]
	println(v2, ok2)

	var ok3 B
	v3, ok3 := m["k"]
	println(v3, ok3)
}

// Output:
// (true main.B) (true main.B)
// (true main.B) true
// (true main.B) (true main.B)
