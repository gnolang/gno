package main

func Search(n int, f func(int) bool) int {
	f(1)
	return 0
}

// TODO: exclude this pattern
func main() {
	for x := 0; x < 2; x++ {
		count := 0
		println(" first: count: ", count)
		Search(1, func(i int) bool { count++; return i >= x })
		println("second: count: ", count)
	}
}

// Output:
// first: count:  0
// second: count:  1
//  first: count:  0
// second: count:  1
