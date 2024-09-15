package main

func main() {
	var ints []int
	ints = append(ints, nil, nil)
	println(ints)
}

// Error:
// main/files/convert5.gno:3:1: cannot convert (undefined) to int
