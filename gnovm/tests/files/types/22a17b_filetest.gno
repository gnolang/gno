package main

type i8 int8
type Int i8

func Add(a, b Int) Int {
	return a + b
}

func main() {
	r := Add(0, i8(1))
	println(r)
}

// Error:
// main/files/types/22a17b_filetest.gno:11: cannot use main.i8 as main.Int without explicit conversion
