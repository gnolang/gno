package main

type IntPtr *int

func main() {
	var a, b int
	a = 1 // Set a to 104
	s := []IntPtr{&a, &b}
	println(*s[0])
}

// Output:
// 1
