package main

func main() {
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

// Error:
// panic: first
// 	panic: second
// 	panic: third
