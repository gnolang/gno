package main

// both typed(different) const
func main() {
	println(true || true)
	println(true || false)
	println(1 == 1 || true)
	println(1 == 1 || false)
}

// Output:
// true
// true
// true
// true
