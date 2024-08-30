package main

func main() {
	m := map[int]int{}
	m[4] = 5
	println(&m[4])
}

// Error:
// main/files/addressable_9a_err.gno:6:10: cannot take address of m<VPBlock(1,0)>[(const (4 int))]
