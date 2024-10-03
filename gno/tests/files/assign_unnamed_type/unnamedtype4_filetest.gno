package main

type nat map[string]int

func main() {
	var a nat

	a = map[string]int{
		"zero": 0,
	}
	b := map[string]int{
		"one": 1,
	}
	println(a)
	println(b)

	a = nat{"zeor": 0}
	b = a
	println(a)
	println(b)
}

// Output:
// (map{("zero" string):(0 int)} main.nat)
// map{("one" string):(1 int)}
// (map{("zeor" string):(0 int)} main.nat)
// map{("zeor" string):(0 int)}
