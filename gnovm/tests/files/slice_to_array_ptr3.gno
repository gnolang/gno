package main

// Slice to array pointer - nil slice to zero-length array pointer is nil.

func main() {
	var s []byte
	p := (*[0]byte)(s)
	println(p == nil)
}

// Output:
// true
