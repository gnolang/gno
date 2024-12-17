package main

type Foo struct {
	num int
}

// different base, two heapItemValue
var s1 = []*Foo{&Foo{1}}
var s2 = []*Foo{&Foo{1}}

func main() {
	println(s1[0] == s2[0])
}
