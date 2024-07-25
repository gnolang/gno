package main

type Error int8

// one untyped const, one typed const
func main() {
	r := Error(1)
	r += 1
	println(r)
}

// Output:
// (2 main.Error)
