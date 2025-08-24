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
// main/assign_range_d.gno:8:2-10:3: index type should be int, but got float32

// TypeCheckError:
// main/assign_range_d.gno:8:6: cannot use index (value of type int) as float32 value in assignment; main/assign_range_d.gno:7:6: declared and not used: value
