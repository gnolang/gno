package main

type Foo struct {
	n int
}

func main() {
	var l interface{} = 1
	println(1 == l)
}

// Output:
// true
