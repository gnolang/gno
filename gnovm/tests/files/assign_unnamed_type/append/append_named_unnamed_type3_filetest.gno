package main

type nat []int

func main() {
	nlist := nat{0}
	ulist := []int{1}

	a := append(nlist, nlist[0], ulist[0])
	println(a)
}

// Output:
// (slice[(0 int),(0 int),(1 int)] main.nat)
