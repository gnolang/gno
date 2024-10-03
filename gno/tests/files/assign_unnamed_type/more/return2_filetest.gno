package main

func x() (int, int) {
	return 1, 2
}

func main() {
	var a, b int
	a, b = x()
	println(a)
	println(b)
}

// Output:
// 1
// 2
