package main

func foo() int {
	{
		return 1
	}
}

func main() {
	println(foo())
}

// Output:
// 1
