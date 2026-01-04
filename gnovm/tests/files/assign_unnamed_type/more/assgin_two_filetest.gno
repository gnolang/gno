package main

type (
	nat  []int
	nat2 []int
)

func main() {
	a := nat{0}
	b := nat2{1}
	a = b

	println(a)
	println(b)
}

// Error:
// main/assgin_two.gno:11:2-7: cannot use main.nat2 as main.nat without explicit conversion

// TypeCheckError:
// main/assgin_two.gno:11:6: cannot use b (variable of slice type nat2) as nat value in assignment
