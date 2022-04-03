package main

type String string

func main() {
	x := "STRING"
	y := String(x)
	println(x + y)
}

// Error:
// incompatible types
