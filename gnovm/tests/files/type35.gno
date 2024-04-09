package main

type IntPtr *int

func main() {
	var a, b int
	a = 1 // Set a to 104
	s := []IntPtr{}
	s = append(s, &a)
	s = append(s, &b)
	println(*s[0])
}

// Output:
// 1
