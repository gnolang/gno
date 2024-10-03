package main

type nat []interface{}

func (n nat) zero() interface{} {
	return []interface{}{0}
}

func (n nat) one() interface{} {
	return nat{1}
}

func main() {
	var a nat
	a = []interface{}{}

	println(a.zero())
	println(a.one())
	println(a.one().(nat).zero())
}

// Output:
// slice[(0 int)]
// (slice[(1 int)] main.nat)
// slice[(0 int)]
