package main

var a = &[]*T{{"hello"}}

type T struct{ name string }

func main() {
	println(*((*a)[0]))
}

// Output:
// (struct{("hello" string)} main.T)
