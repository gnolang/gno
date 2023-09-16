package main

func main() {
	var x string
	y := "Hello"
	if x < y {
		println("x < y")
	}
	if y < x {
		println("y < x")
	}
	if x == "" {
		println("x empty")
	}
	if y == "" {
		println("y empty")
	}
}

// Output:
// x < y
// x empty
