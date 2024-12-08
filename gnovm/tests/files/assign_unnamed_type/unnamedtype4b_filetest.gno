package main

type nat map[string]int

func (n nat) zero() nat {
	return map[string]int{"zero": 0}
}

func (n nat) one() map[string]int {
	return nat{"one": 1}
}

func main() {
	var a nat
	a = map[string]int{"zero": 0}

	println(a.zero())
	println(a.zero().zero())
	println(a.zero().one())
	println(a.one())
}

// Output:
// (map{("zero" string):(0 int)} main.nat)
// (map{("zero" string):(0 int)} main.nat)
// map{("one" string):(1 int)}
// map{("one" string):(1 int)}
