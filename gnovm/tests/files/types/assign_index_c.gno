package main

func main() {
	s := []int{1, 2, 3}

	var s1 string
	s1 = s[0]
	println(s)
}

// Error:
// main/files/types/assign_index_c.gno:7:2: cannot use int as string

// TypeCheckError:
// main/files/types/assign_index_c.gno:7:7: cannot use s[0] (variable of type int) as string value in assignment; main/files/types/assign_index_c.gno:6:6: declared and not used: s1
