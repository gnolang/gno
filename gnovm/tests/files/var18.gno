package main

func main() {
	println(a)
	println(b)
}

func r2() (int, int) { return 1, 2 }

var a, b int = r2()

// Output:
// 1
// 2
