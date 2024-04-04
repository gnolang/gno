package main

type nat [1]int

func (n nat) zero() interface{} {
	return [1]int{0}
}

func (n nat) one() interface{} {
	return nat{1}
}

func main() {
	var a nat
	a = [1]int{}

	println(a.zero())
	println(a.one())
	println(a.one().(nat).zero())
}

// Output:
// array[(0 int)]
// (array[(1 int)] main.nat)
// array[(0 int)]
