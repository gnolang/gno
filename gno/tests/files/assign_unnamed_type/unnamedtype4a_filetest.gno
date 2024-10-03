package main

type nat map[string]int

func (n nat) zero(num map[string]int) {
	println(num)
}

func (n nat) one(num nat) {
	println(num)
}

func main() {
	var a nat
	a = map[string]int{
		"zero": 0,
	}

	a.zero(nat{"zero": 0})
	a.one(map[string]int{
		"zero": 1,
	})
}

// Output:
// map{("zero" string):(0 int)}
// (map{("zero" string):(1 int)} main.nat)
