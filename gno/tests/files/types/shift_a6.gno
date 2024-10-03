package main

func main() {
	r := int(1)
	println(r << 5)
	println(r)
	r <<= 5
	println(r)
}

// Output:
// 32
// 1
// 32
