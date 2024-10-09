package main

func main() {
	a := 1
	r := a << "hello" // NOTE: go vet would fail, but still process
	println(r)
}

// Error:
// main/files/types/shift_a11.gno:5:7: cannot convert (const ("hello" <untyped> string)) to type uint
