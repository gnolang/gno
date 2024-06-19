package main

func main() {
	a := 1
	r := a << "hello" // NOTE: go vet would fail, but still process
	println(r)
}

// Error:
// main/files/types/shift_a11.gno:3: cannot convert StringKind to UintKind
