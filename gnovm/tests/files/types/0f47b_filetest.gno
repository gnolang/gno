package main

type integer *int

// nil can be cast to pointer
func main() {
	println(integer(nil) == nil)
}

// Output:
// true
