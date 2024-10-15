package main

func main() {
	s1 := []int{1, 2}
	i0 := &s1[0]
	// new array allocated, so they have different base array
	s1 = append(s1, 3)
	ii0 := &s1[0]
	println(i0 == ii0)
}

// Output:
// false
