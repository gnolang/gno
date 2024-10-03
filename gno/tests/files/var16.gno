package main

func main() {
	x := 1
	{
		x := (func() int { x := x + 100; return x + 10000 })()
		println(x)
	}
}

// Output:
// 10101
