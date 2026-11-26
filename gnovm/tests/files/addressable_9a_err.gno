package main

func main() {
	m := map[int]int{}
	m[4] = 5
	println(&m[4])
}

// TypeCheckError:
// main/addressable_9a_err.gno:6:11: invalid operation: cannot take address of m[4] (map index expression of type int)

// Error:
// main/addressable_9a_err.gno:6:10-15: invalid operation: cannot take address of m<VPBlock(1,0)>[(const (4 int))] (value of type int)
