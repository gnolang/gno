// Named int as range key: the index is assigned, so named-type assignability
// applies — int is not assignable to a defined type without conversion.
package main

type I int

func main() {
	var i I
	for i = range []string{"a", "b"} {
	}
	println(i)
}

// Error:
// main/assign_range_j.gno:9:2-10:3: cannot use int as main.I without explicit conversion

// TypeCheckError:
// main/assign_range_j.gno:9:6: cannot use i (value of type int) as I value in assignment
