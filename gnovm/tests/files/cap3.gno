package main

func main() {
	slice := make([]int, 3, 5)
	println(cap(slice))
}

// Output:
// 5
