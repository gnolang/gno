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
// &0x14000189110.(*main.foo)
