package main

type Error int8

// one untyped const, one typed const
func main() {
	println(1 % Error(1))
	println(Error(1) % 1)
}

// Output:
// (0 main.Error)
// (0 main.Error)
