package main

type Foo struct {
	val int
}

func main() {
	a := &Foo{3}
	println(a.val)
}

// Output:
// 3
