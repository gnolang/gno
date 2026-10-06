package main

// ref of a struct variable.
// xv.TV.T is main.MyStruct, which is correct.

type MyStruct struct {
	A int
}

func main() {
	var s MyStruct
	s.A = 1
	p := &s
	p.A = 2
	println(s)
}

// Output:
// (struct{(2 int)} main.MyStruct)
