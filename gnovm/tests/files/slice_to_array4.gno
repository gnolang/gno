package main

// Slice to array conversion with non-byte types (int).

func main() {
	s := []int{10, 20, 30}
	a := [3]int(s)
	println(a[0] + a[1] + a[2])
}

// Output:
// 60
