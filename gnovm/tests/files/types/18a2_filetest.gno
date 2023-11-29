package main

func main() {
	r := int(1)
	r &= int8(1)
	println(r)
}

// Error:
// main/files/types/18a2_filetest.gno:5: cannot use int8 as int
