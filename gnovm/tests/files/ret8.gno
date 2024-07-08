package main

func test() int {
	x := [4]int{1, 2, 3, 4}
	for range x {
		return 2
	}
	println("after for")
	return 1
}

func main() {
	println(test())
}

// Output:
// 2
