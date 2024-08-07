package main

func main() {
	a := 1
	r := a << "hello" // NOTE: go vet would fail, but still process
	println(r)
}

// Error:
// main/files/types/shift_a11.gno:5:7: invalid operation: invalid shift count: (const ("hello" <untyped> string))
