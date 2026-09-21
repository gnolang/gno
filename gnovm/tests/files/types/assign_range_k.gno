// Named rune as string range value: the value is assigned, so named-type
// assignability applies.
package main

type R rune

func main() {
	var r R
	for _, r = range "ab" {
	}
	println(r)
}

// Error:
// main/assign_range_k.gno:9:2-10:3: cannot use int32 as main.R without explicit conversion

// TypeCheckError:
// main/assign_range_k.gno:9:9: cannot use r (value of type rune) as R value in assignment
