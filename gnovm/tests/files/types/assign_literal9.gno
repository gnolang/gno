package main

type foo struct {
	a int
}

func main() {
	s := &foo{}
	s.a = 1
	println(s)
}

// Output:
// &0x140002f85a0.(*main.foo)
