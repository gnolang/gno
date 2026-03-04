package main

var sink int

func main() {
	var y int
	const s = 1
	{
		y = (1.0 << s) << s 
	}
	sink = y
	println(sink)
}

// Output:
// 4