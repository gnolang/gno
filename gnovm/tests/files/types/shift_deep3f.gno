package main

var sink int

func main() {
	var y int
	// build deep left-leaning chain: (((y<<s)<<s)<<s)...
	{
		y = (1.0 << 1) << 1 
	}
	sink = y
	println(sink)
}

// Output:
// 4