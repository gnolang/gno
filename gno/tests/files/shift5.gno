package main

func maxDepth(n int) int {
	var depth int
	for i := n; i > 0; i >>= 1 {
		depth++
	}
	return depth * 2
}

func main() {
	println(maxDepth(10))
}

// Output:
// 8
