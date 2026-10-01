// `nil` as the range value operand: its static type is nil, aliasing with a
// blank operand at the type level — only the syntactic lhs check catches it.
package main

func main() {
	m := map[string]int{"a": 1}
	var k string
	for k, nil = range m {
	}
	_ = k
	println("ok")
}

// Error:
// main/assign_range_h.gno:8:2-9:3: cannot assign to (const (undefined))

// TypeCheckError:
// main/assign_range_h.gno:8:9: cannot assign to nil (neither addressable nor a map index expression)
