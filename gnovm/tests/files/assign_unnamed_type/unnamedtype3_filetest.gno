package main

type nat struct {
	num int
}

func main() {
	var a nat
	a = struct{ num int }{0}
	b := struct{ num int }{1}

	println(a)
	println(b)
	a = nat{0}
	b = a

	println(a)
	println(b)
}

// Output:
// (struct{(0 int)} main.nat)
// struct{(1 int)}
// (struct{(0 int)} main.nat)
// struct{(0 int)}
