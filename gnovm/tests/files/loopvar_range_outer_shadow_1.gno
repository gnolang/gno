package main

// Range loop var with same name as an outer var: outer var should be unchanged.
func main() {
	i := 99
	for i, v := range []int{10, 20, 30} {
		_ = v
		_ = i
	}
	println("outer i:", i) // 99
}

// Output:
// outer i: 99
