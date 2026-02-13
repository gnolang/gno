package main

// Slice to array pointer from sub-slice - tests offset in pointer.
// Modifying p[0] should affect s[1], not s[0].

func main() {
	s := make([]byte, 2, 4)
	s[0] = 10
	s[1] = 20
	p := (*[1]byte)(s[1:])
	p[0] = 99
	println(s[0], s[1], p[0])
}

// Output:
// 10 99 99
