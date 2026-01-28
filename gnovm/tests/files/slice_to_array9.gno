package main

// Slice to array conversion - mismatched element types should fail.

func main() {
	s := []int{1, 2, 3}
	a := [3]byte(s)
	_ = a
}

// Error:
// main/slice_to_array9.gno:7:7-17: cannot convert s<VPBlock(1,0)> (of type []int) to type [3]uint8

// TypeCheckError:
// main/slice_to_array9.gno:7:15: cannot convert s (variable of type []int) to type [3]byte
