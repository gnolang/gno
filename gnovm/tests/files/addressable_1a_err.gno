package main

func main() {
	_ = &[1]int{1}[0]
}

// TypeCheckError:
// main/files/addressable_1a_err.gno:4:7: invalid operation: cannot take address of [1]int{…}[0] (value of type int)
