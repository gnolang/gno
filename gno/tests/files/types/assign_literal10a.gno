package main

type foo struct {
	a int
}

var n int

func (f foo) add() *int { return &n }

func main() {
	s := &foo{}
	s.add() = 1
	println((*s).a)

	println(n)
}

// Error:
// main/files/types/assign_literal10a.gno:13:2: cannot assign to s<VPBlock(1,0)>.add()
