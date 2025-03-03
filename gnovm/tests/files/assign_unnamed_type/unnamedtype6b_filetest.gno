package main

type nat []any

func (n nat) zero() nat {
	return []any{0}
}

func (n nat) one() []any {
	return nat{1}
}

func main() {
	var a nat
	a = []any{}

	println(a.zero())
	println(a.zero().zero())
	println(a.zero().one())
	println(a.one())
}

// Output:
// (slice[(0 int)] main.nat)
// (slice[(0 int)] main.nat)
// slice[(1 int)]
// slice[(1 int)]
