package main

func main() {
	var ints []int
	ints = append(ints, nil, nil)
	println(ints)
}

// Error: cannot convert (undefined) to int
