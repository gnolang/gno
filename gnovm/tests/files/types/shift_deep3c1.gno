package main


func main() {
	var s int = 1
	var y int
	// build deep left-leaning chain: (((y<<s)<<s)<<s)...
	{
		y = (1.0 << s) + int(1)
	}
	println(y)
}

// Output:
// 3
