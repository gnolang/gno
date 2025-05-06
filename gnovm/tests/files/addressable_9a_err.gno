package main

func main() {
	m := map[int]int{}
	m[4] = 5
	println(&m[4])
}

// Output:
// &(5 int)

// TypeCheckError:
// main/files/addressable_9a_err.gno:6:11: invalid operation: cannot take address of m[4] (map index expression of type int)
