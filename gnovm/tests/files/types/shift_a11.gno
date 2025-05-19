package main

func main() {
	a := 1
	r := a << "hello" // NOTE: go vet would fail, but still process
	println(r)
}

// Error:
// main/shift_a11.gno:5:7-19: cannot convert (const ("hello" <untyped> string)) to type uint

// TypeCheckError:
// main/shift_a11.gno:5:12: cannot convert "hello" (untyped string constant) to type uint
