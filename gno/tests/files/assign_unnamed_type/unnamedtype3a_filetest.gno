package main

type nat struct {
	num int
}

func (n nat) zero(num struct{ num int }) {
	println(num)
}

func (n nat) one(num nat) {
	println(num)
}

func main() {
	var a nat
	a = struct{ num int }{0}

	a.zero(nat{0})
	a.one(struct{ num int }{1})
}

// Output:
// struct{(0 int)}
// (struct{(1 int)} main.nat)
