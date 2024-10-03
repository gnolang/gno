package main

type T uint

func main() {
	type myint int
	i := myint(1)
	println(i)
}

// Output:
// (1 main.myint)
