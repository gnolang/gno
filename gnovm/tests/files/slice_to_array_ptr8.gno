package main

// Empty non-nil slice to zero-length array pointer is non-nil.

func main() {
	u := make([]byte, 0)
	p := (*[0]byte)(u)
	println(p != nil)
}

// Output:
// true
