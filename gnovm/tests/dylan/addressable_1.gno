package main

func main() {
	// Array pointers are addressable.
	println(&getArrPtr1()[0])
	println(&getArrPtr2()[0])
	println(&getArrPtr3()[0])
	println(&new([1]int)[0])

	// Array pointers are slicable.
	println(getArrPtr1()[:])
	println(getArrPtr2()[:])
	println(getArrPtr3()[:])
	println(new([1]int)[:])
}

func getArrPtr1() *[1]int {
	return &[1]int{1}
}

func getArrPtr2() *[1]int {
	a := [1]int{2}
	return &a
}

func getArrPtr3() *[1]int {
	return new([1]int)
}

// Output:
// &(1 int)
// &(2 int)
// &(0 int)
// &(0 int)
// slice[(1 int)]
// slice[(2 int)]
// slice[(0 int)]
// slice[(0 int)]
