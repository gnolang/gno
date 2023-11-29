package main

type Foo struct {
	val int
}

func main() {
	a := Foo{3}
	b := &a
	println(b.val)
}

// Output:
// 3
