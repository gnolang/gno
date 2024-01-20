package main

type s1 struct {
	foo int
}

func main() {
	println(s1{foo: 1} == 1)
}

// Error:
// main/files/types/22a18_filetest.gno:8: cannot use <untyped> bigint as struct{foo int}
