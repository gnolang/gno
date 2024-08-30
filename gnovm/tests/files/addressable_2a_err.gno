package main

func main() {
	_ = &[]int{1}[:]
}

// Error:
// main/files/addressable_2a_err.gno:4:6: cannot take address of [](const-type int){(const (1 int))}[:]
