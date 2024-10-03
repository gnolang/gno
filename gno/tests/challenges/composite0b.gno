package main

// NOTE: this requires implementing String2() to
// prevent infinite recursion.

var a = &[]*T{}

type T struct{ name string }

func main() {
	println(a)
}

// Output:
// &struct{("" string)}
