package main

type nat []int

func main() {
	var nlist = nat{0}
	var ulist = []int{1}
	var n int = 3

	a := append(nlist, n)
	b := append(ulist, n)

	println(a)
	println(b)

	a = append(ulist, n)
	b = append(nlist, n)

	c := append(nlist, ulist...)
	d := append(ulist, nlist...)

	println(a)
	println(b)
	println(c)
	println(d)

}

// Output:
// (slice[(0 int),(3 int)] main.nat)
// slice[(1 int),(3 int)]
// (slice[(1 int),(3 int)] main.nat)
// slice[(0 int),(3 int)]
// (slice[(0 int),(1 int)] main.nat)
// slice[(1 int),(0 int)]
