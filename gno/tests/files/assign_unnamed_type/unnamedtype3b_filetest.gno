package main

type nat struct {
	num int
}

func (n nat) zero() nat {
	return struct{ num int }{0}
}

func (n nat) one() struct{ num int } {
	return nat{1}
}

func main() {
	var a nat
	a = struct{ num int }{}

	println(a.zero())
	println(a.zero().zero())
	println(a.zero().one())
	println(a.one())
}

// Output:
// (struct{(0 int)} main.nat)
// (struct{(0 int)} main.nat)
// struct{(1 int)}
// struct{(1 int)}
