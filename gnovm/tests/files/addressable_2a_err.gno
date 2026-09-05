package main

func main() {
	_ = &[]int{1}[:]
}

// Error:
// main/addressable_2a_err.gno:4:6-18: invalid operation: cannot take address of (const-type []int){(const (1 int))}[:] (value of type []int)

// TypeCheckError:
// main/addressable_2a_err.gno:4:7: invalid operation: cannot take address of []int{…}[:] (value of type []int)
