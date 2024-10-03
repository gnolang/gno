package main

type foo struct {
	a int
}

var n int

func (f foo) add() *int { return &n }

func main() {
	s := &foo{}
	*(s.add()) = 1
	println((*s).a)

	println(n)
}

// Output:
// 0
// 1
