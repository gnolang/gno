package main

func main() {
	var ints []int
	ints = append(ints, nil, nil)
	println(ints)
}

// error: cannot convert (undefined) to int
