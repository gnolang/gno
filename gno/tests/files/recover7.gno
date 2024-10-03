package main

func main() {
	defer func() {
		println("1.", recover())
		println("2.", recover())
		println("3.", recover())
	}()
	f()
}

func f() {
	defer func() {
		panic("third")
	}()
	defer func() {
		panic("second")
	}()
	panic("first")
}

// Output:
// 1. third
// 2. undefined
// 3. undefined
