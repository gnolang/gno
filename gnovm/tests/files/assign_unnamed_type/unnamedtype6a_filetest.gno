package main

type nat []any

func (n nat) zero(num []any) {
	println(num)
}

func (n nat) one(num nat) {
	println(num)
}

func main() {
	var a nat
	a = []any{}
	println(a)
	a.zero(nat{0})
	a.one([]any{1})
}

// Output:
// (slice[] main.nat)
// slice[(0 int)]
// (slice[(1 int)] main.nat)

