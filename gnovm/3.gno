package main

type Foo struct {
	name string
}

var root interface{}

var f = &Foo{name: "a"}

func init() {
	root = f
}

func main() {
	f.name = "b"

	println(root == f)
}
