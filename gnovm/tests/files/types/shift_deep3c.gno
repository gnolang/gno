package main

func main() {
	var s int = 1
	var y int8
	// build deep left-leaning chain: (((y<<s)<<s)<<s)...
	{
		y = (1 << s) << s 
	}
	println(y)
}

// Output:
// 4
