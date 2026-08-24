package main

type T struct {
	_ int
	X int
}

type Outer struct {
	Inner T
}

func main() {
	println(T{1, 2} == T{3, 2})
	println(T{1, 2} == T{1, 3})
	println(Outer{T{1, 2}} == Outer{T{3, 2}})
}

// Output:
// true
// false
// true
