package main

func main() {
	var x map[int]string
	var y map[int]int
	x = y
	_ = x
}

// Error:
// main/assign_map_value_type.gno:6:2-7: cannot use map[int]int as map[int]string

// TypeCheckError:
// main/assign_map_value_type.gno:6:6: cannot use y (variable of type map[int]int) as map[int]string value in assignment
