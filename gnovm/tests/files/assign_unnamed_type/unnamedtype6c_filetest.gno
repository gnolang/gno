package main

type nat []any

func (n nat) zero() any {
	return []any{0}
}

func (n nat) one() any {
	return nat{1}
}

func main() {
	var a nat
	a = []any{}

	println(a.zero())
	println(a.one())
	println(a.one().(nat).zero())
}

// Output:
// slice[(0 int)]
// (slice[(1 int)] main.nat)
// slice[(0 int)]
