package main

var sink int

func main() {
	var s int = 1
	var y int
	{
		y = 1.0 << s
	}
	sink = y
	println(sink)
}

// Output:
// 2
