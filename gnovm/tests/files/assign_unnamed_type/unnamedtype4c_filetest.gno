package main

type nat map[string]int

func (n nat) zero() any {
	return map[string]int{"zero": 0}
}

func (n nat) one() any {
	return nat{"one": 1}
}

func main() {
	var a nat
	a = map[string]int{"zero": 0}

	println(a.zero())
	println(a.one())
	println(a.one().(nat).zero())
}

// Output:
// map{("zero" string):(0 int)}
// (map{("one" string):(1 int)} main.nat)
// map{("zero" string):(0 int)}
