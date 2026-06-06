package main

func main() {
	var s int = 1
	var y int8
	{
		y = (1 << s) << s 
	}
	println(y)
}

// Output:
// 4
