package main

// Slice to array pointer - panic when slice is too short.

func main() {
	s := []byte{1, 2}
	p := (*[5]byte)(s)
	_ = p
}

// Error:
// cannot convert slice with length 2 to pointer to array with length 5
