// Range over a declared (named) map type: operand checks must apply to the
// base type, not skip on *DeclaredType.
package main

type M map[string]int

func main() {
	m := M{"a": 1}
	var k, v int
	for k, v = range m {
	}
	println(k, v)
}

// Error:
// main/assign_range_i.gno:10:2-11:3: cannot use string as int

// TypeCheckError:
// main/assign_range_i.gno:10:6: cannot use k (value of type string) as int value in assignment
