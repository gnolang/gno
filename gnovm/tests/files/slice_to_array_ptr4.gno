package main

// Slice to array pointer - nil slice to non-zero array pointer panics.

func main() {
	var s []byte
	p := (*[1]byte)(s)
	_ = p
}

// Error:
// cannot convert slice with length 0 to pointer to array with length 1
