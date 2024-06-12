package main

// both untyped const
func main() {
	println(1.0 == 1)
	println(1.0 == 0)
	println(float32(1.0) == 0)
}

// Output:
// true
// false
// false
