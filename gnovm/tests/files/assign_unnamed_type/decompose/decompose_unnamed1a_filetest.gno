package main

type nat []int

func x() (nat, []int) {
	a := nat{1}
	b := []int{2}
	return a, b
}

func main() {
	var u1 []int
	var n2 nat

	_, n2 = x()
	// _tmp1, _tmp_2  := x()
	// _, u2 = _tmp1, _tmp_2

	println(u1)
	println(n2)

}

// Output:
// (nil []int)
// (slice[(2 int)] main.nat)
