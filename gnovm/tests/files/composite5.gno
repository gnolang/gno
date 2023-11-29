package main

type T struct {
	m uint16
}

var t = T{1<<2 | 1<<3}

func main() {
	println(t)
}

// Output:
// (struct{(12 uint16)} main.T)
