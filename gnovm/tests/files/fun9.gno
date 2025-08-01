package main

type T uint

func main() {
	type myint int
	i := myint(1)
	println(i)
}

// Output:
// (1 main[main/fun9.gno:5:1-9:2].myint)
