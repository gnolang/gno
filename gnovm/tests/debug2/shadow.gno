package main

func main() {
	var x int

	if true {
		x += 2
		println("inner x: ", x)
	}

	println("outer x:", x)
}

// Output:
// inner x:  2
// outer x: 2
