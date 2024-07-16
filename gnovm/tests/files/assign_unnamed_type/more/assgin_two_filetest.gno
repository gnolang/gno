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
// main/files/assign_unnamed_type/more/assgin_two_filetest.gno:11:2: cannot use main.nat2 as main.nat without explicit conversion
