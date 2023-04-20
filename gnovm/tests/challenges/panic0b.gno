package main

func main() {
	f()
}

func f() {
	defer func() {
		panic("second")
	}()
	panic("first")
}

// Error:
// first
//	second
