package main

func main() {
	f := func() { println("Hello, World!") }
	println(&f)
}

// Output:
// &(func()(){...} func()())
