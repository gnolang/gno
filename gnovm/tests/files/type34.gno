package main

type BytePtr *byte

func main() {
	bs := []byte("hello")
	var p BytePtr = &bs[0]
	println(*p)
}

// Output:
// 104
