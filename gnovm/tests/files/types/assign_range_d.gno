package main

func main() {
	s := "hello"

	var index float32
	var value rune
	for index, value = range s {
		println(index)
	}
}

// Error:
// main/files/types/assign_range_d.gno:8:2: index type should be int, but got float32
