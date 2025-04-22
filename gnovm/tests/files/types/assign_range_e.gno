package main

func main() {
	s := "hello"

	var index int
	var value int
	for index, value = range s {
		println(index)
		println(value)
	}
}

// Error:
// main/files/types/assign_range_e.gno:8:2: value type should be int32, but got int

// TypeCheckError:
// main/files/types/assign_range_e.gno:8:13: cannot use value (value of type rune) as int value in assignment
