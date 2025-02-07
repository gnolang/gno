package main

func main() {
	var s uint = 33
	var u2 = uint64(1<<s + 1.0)
	println(u2)
}

// Output:
// 8589934593
