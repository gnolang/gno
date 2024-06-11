package main

type nat []int

func main() {
	var nlist = nat{0}
	var ulist = []int{1}

	copy(nlist, ulist)

	println(nlist)
	println(ulist)

	nlist = nat{0}
	ulist = []int{1}
	copy(ulist, nlist)
	println(nlist)
	println(ulist)
}

// Output:
// (slice[(1 int)] main.nat)
// slice[(1 int)]
// (slice[(0 int)] main.nat)
// slice[(0 int)]
