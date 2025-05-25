package main

type Foo struct {
	n int
}

func main() {
	var l any = 1
	println(1 == l)
}

// Output:
// true
