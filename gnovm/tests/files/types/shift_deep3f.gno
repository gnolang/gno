package main

var sink int

func main() {
	var y int
	{
		y = (1.0 << 1) << 1 
	}
	sink = y
	println(sink)
}

// Output:
// 4