package main

type nat struct {
	num int
}

func (n nat) zero() any {
	return struct{ num int }{0}
}

func (n nat) one() any {
	return nat{1}
}

func main() {
	var a nat
	a = struct{ num int }{}

	println(a.zero())
	println(a.one())
	println(a.one().(nat).zero())
}

// Output:
// struct{(0 int)}
// (struct{(1 int)} main.nat)
// struct{(0 int)}
