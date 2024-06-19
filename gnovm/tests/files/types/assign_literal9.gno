package main

type foo struct {
	a int
}

func main() {
	s := &foo{}
	s.a = 1
	println(s.a)
}

// Output:
// 1
