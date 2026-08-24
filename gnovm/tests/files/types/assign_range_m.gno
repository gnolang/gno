// Range over a named pointer-to-array type: the source type must be resolved
// through baseOf (DeclaredType -> *PointerType) and then unwrapped to the array
// before the value target's type is checked.
package main

type P *[2]int

func main() {
	a := [2]int{1, 2}
	var p P = &a
	var k int
	var s string
	for k, s = range p {
	}
	_, _ = k, s
}

// Error:
// main/assign_range_m.gno:13:2-14:3: cannot use int as string

// TypeCheckError:
// main/assign_range_m.gno:13:9: cannot use s (value of type int) as string value in assignment
