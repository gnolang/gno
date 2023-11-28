package main

func main() {
	a := 6 // a is 6 (binary 110)
	b := 3 // b is 3 (binary 011)
	a &= b // Now a is 2 (binary 010)

	println(a)
}

// Output:
// 2
