package main

func main() {
	// Test slice truncation
	s := make([]int, 300)
	println(s)

	// Test array truncation
	var a [300]int
	println(a)

	// Test map truncation
	m := make(map[int]int)
	for i := 0; i < 300; i++ {
		m[i] = i
	}
	println(m)
}

// Output:
// slice[...(300 elements)]
// array[...(300 elements)]
// map{...(300 entries)}
