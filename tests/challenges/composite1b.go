package main

var a = &[]*T{{}}

type T struct{ name string }

func main() {
	println((*a)[0])
}

// Output:
// &struct{("" string)}
