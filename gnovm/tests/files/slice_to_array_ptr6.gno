package main

// Slice to array pointer - non-byte types (int).

func main() {
	s := []int{10, 20, 30}
	p := (*[3]int)(s)
	p[0] = 99
	println(s[0], p[0])
}

// Output:
// 99 99
