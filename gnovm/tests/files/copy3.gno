package main

func main() {
	var dst []byte // nil
	copy(dst, []byte("123"))
	println(dst)
}

// Output:
// (nil []uint8)
