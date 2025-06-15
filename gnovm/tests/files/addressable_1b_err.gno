package main

func main() {
	_ = [1]int{1}[:]
}

// TypeCheckError:
// main/addressable_1b_err.gno:4:6: invalid operation: cannot slice [1]int{…} (value of type [1]int) (value not addressable)
