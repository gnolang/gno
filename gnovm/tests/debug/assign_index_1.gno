package main

func main() {
	s := []int{1, 2, 3}

	var s1 string
	s1 = s[0]
	println(s)
}

// Error:
// main/debug/assign_index_1.gno:7: cannot use int as string
