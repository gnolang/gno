package main

func main() {
	var v uint64 = 1
	b := byte(v >> 8)
	println(b)
}

// Output:
// 0
