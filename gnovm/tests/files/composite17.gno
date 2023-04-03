package main

type Struct struct {
	Field string
}

func main() {
	var x []*Struct = []*Struct{
		{"foobar"},
		{"barbaz"},
	}
	println(x[0].Field)
	println(x[1].Field)
}

// Output:
// foobar
// barbaz
