package main

type Foo interface {
	Len() int
}

type myIntSlice []int

func (s myIntSlice) Len() int {
	return len(s)
}

func main() {
	var a Foo
	var b interface{ Len() int }

	var i myIntSlice

	println(a)

	i = myIntSlice{1}
	a = i

	println(a)

	b = interface{ Len() int }(a)

	a = b

	println(a)

	println(b)
}

// Output:
// undefined
// (slice[(1 int)] main.myIntSlice)
// (slice[(1 int)] main.myIntSlice)
// (slice[(1 int)] main.myIntSlice)
