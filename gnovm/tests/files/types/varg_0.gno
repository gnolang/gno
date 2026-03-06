package main

func main() {
	var a uint = 1

	b := make([]byte, 1.0<<a)

	println(b)
}


// Output:
// slice[0x0000]