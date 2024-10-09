package main

func main() {
	var s uint = 33
	var u2 = 1.0<<s + 1 + int(1)
	println(u2)
}

// Output:
// 8589934594
