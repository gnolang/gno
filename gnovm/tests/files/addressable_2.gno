package main

func main() {
	// Slices are addressable because the underlying array is addressable
	// after slice initialization.
	println(&[]int{1}[0])
	println(&getSlice1()[0])
	println(&getSlice2()[0])
	println(&[]int{1})

	a := []int{1}
	println(&append(a, 1, 2, 3, 4, 5)[5])

	println([]int{1}[:])
	println(getSlice1()[:])
	println(getSlice2()[:])

	b := []int{1}
	println(append(b, 1, 2, 3, 4, 5)[:])
}

func getSlice1() []int {
	return []int{2}
}

func getSlice2() []int {
	s := []int{3}
	return s
}

// Output:
// &(1 int)
// &(2 int)
// &(3 int)
// &(slice[(1 int)] []int)
// &(5 int)
// slice[(1 int)]
// slice[(2 int)]
// slice[(3 int)]
// slice[(1 int),(1 int),(2 int),(3 int),(4 int),(5 int)]
