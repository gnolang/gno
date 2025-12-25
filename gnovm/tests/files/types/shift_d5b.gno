package main

func main() {
	x := 11
	var y int
	y = 1.0 << x // no const
	println(y)
}

// Output:
// 2048