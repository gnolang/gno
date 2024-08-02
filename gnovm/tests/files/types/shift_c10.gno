package main

func check(a uint, b float32) {
	println(a, b)
}

// case of expr in type call
func main() {
	check(1<<2, 1>>2)
}

// Output:
// 4 0
