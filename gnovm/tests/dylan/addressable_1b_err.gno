package main

func main() {
	_ = [1]int{1}[:]
}

// Error:
// main/dylan/addressable_1b_err.gno:4:6: cannot take address of [(const (1 int))](const-type int){(const (1 int))}
