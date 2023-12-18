package main

type Int int8

func Add(a, b Int) Int {
	return a + b
}

func main() {
	r := Add(0, 1)
	println(r)
}

// Output:
// (1 main.Int)
