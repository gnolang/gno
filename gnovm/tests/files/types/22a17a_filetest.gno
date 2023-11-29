package main

type Int int8

func Add(a, b Int) Int {
	return a + b
}

func main() {
	r := Add(0, int8(1))
	println(r)
}

// Error:
// main/files/types/22a17a_filetest.gno:10: cannot use int8 as main.Int without explicit conversion
