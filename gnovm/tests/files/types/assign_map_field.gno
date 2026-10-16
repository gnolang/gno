package main

type X struct{ Array [8]int }

func main() {
	xm := map[int]X{1: X{}}
	xm[1].Array[1] = 123 // a map element is not addressable
}

// Error:
// main/assign_map_field.gno:7:2-22: cannot assign to xm[1].Array[1] (neither addressable nor a map index expression)

// TypeCheckError:
// main/assign_map_field.gno:7:2: cannot assign to xm[1].Array[1] (neither addressable nor a map index expression)
