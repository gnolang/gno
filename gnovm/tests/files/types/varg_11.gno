package main

func main() {
	n32 := int32(3)
	s := make([]int, n32)
	println(len(s))

	n64 := int64(2)
	m := make([]byte, n64)
	println(len(m))
}

// Output:
// 3
// 2
