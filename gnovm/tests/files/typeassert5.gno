package main

type A interface {
	Do(i int)
}

type B interface {
	Do(i int)
}

type C struct{}

func (c C) Do(i int) {}

func AcceptA(a A) {
	AcceptB(a)
}

func AcceptB(b B) {
	if _, ok := b.(A); ok {
		println("ok")
	}
}

func main() {
	AcceptA(C{})
}

// Output:
// ok
