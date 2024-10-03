package main

type nat []interface{}

func (n nat) zero() nat {
	return []interface{}{0}
}

func (n nat) one() []interface{} {
	return nat{1}
}

func main() {
	var a nat
	a = []interface{}{}

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
