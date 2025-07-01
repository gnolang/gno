package main

func main() {
	s := make([][2]string, 1)  // Slice with length 1
	s[0] = [2]string{"a", "b"} // Assign value to s[0]
	const r = len(s[0])
	println(r) // Prints: 2
}

// Output:
// 2
