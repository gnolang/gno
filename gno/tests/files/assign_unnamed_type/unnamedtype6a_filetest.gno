package main

type nat []interface{}

func (n nat) zero(num []interface{}) {
	println(num)
}

func (n nat) one(num nat) {
	println(num)
}

func main() {
	var a nat
	a = []interface{}{}
	println(a)
	a.zero(nat{0})
	a.one([]interface{}{1})
}

// Output:
// (slice[] main.nat)
// slice[(0 int)]
// (slice[(1 int)] main.nat)
