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

	u1, _ = x()
	// .tmp1, .tmp_2  := x()
	// u1, _ = .tmp1, .tmp_2

	println(u1)
	println(n2)

}

// Output:
// slice[(1 int)]
// (nil main.nat)
