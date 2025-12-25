package main

var sink int

func main() {
	var s int = 1
	var y int
	// build deep left-leaning chain: (((y<<s)<<s)<<s)...
	{
		y = 1.0 << s
	}
	sink = y
	println(sink)
}

// Output:
// 2
