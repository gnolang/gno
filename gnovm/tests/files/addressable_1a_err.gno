package main

func main() {
	_ = &[1]int{1}[0]
}

// TypeCheckError:
// main/addressable_1a_err.gno:4:7: invalid operation: cannot take address of [1]int{…}[0] (value of type int)

// Error:
// main/addressable_1a_err.gno:4:6-19: invalid operation: cannot take address of (const-type [1]int){(const (1 int))}[(const (0 int))] (value of type int)
