package main

func p(i int) {
	defer func() {
		panic("here")
	}()
	if i == 4 {
		panic("here")
	}
	p(i + 1)
}

func main() {
	p(0)
}

// Stacktrace:
// panic: here

// Error:
// here
// 	here
// 	here
// 	here
// 	here
// 	here
