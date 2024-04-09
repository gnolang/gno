package main

type Foo struct {
	age  int
	name string
}

func main() {
	f := Foo{age: 1, name: "hello"}
	println(f.age)
}

// Output:
// 1
