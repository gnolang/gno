package main

// Slice to array pointer - non-nil slice to zero-length array pointer is non-nil.

func main() {
	s := []byte{1, 2, 3}
	p := (*[0]byte)(s)
	println(p == nil)
}

// Output:
// false
