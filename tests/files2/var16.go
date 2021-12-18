package main

func main() {
	var x = 1
	{
		var x = (func() int { var x = x + 100; return x + 10000 })()
		println(x)
	}
}

// Output:
// 10101
