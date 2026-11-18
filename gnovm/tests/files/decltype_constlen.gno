package main

// An array length can depend on a constant converted to another type of
// the group. E's length needs T built, not just reserved, so N's
// predefinition builds T first.

type E [N]int

const N = T(3)

type T int

func main() {
	var e E
	println(len(e))
}

// Output:
// 3
