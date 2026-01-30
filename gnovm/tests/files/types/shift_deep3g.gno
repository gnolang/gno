package main

var sink int

func main() {
	var y int
	const s = 1
	// build deep left-leaning chain: (((y<<s)<<s)<<s)...
	{
		y = (1.0 << s) << s 
	}
	sink = y
	println(sink)
}

// Output:
// 4