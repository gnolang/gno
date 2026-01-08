package main

var sink int

func main() {
	var x int = 1
	var s int = 1
	var y int
	// build deep left-leaning chain: (((y<<s)<<s)<<s)...
	{
		y = (x << s) << s 
	}
	sink = y
	println(sink)
}

// Output:
// 4
