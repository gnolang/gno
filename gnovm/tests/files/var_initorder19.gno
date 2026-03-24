package main

var A = func() int {
	var (
		local = B
		_     = 123 // just to avoid `fmt` simplifying this into :=.
	)
	return local
}()

var B = 42

func main() {
	println(A)
}

// Output:
// 42
