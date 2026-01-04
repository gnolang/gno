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
// main/assign_range_e.gno:8:2-11:3: value type should be int32, but got int

// TypeCheckError:
// main/assign_range_e.gno:8:13: cannot use value (value of type rune) as int value in assignment
