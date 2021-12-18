package main

func test() int {
	for _, dir := range [...]int{1, 2, 3, 4} {
		if dir > 2 {
			return 2
			println("after if")
		}
	}
	println("after for")
	return 1
}

func main() {
	println(test())
}

// Output:
// 2
