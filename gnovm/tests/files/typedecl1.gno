package main

// Type declarations are valid in any block, not only at package level and
// directly inside a function body.
func main() {
	{
		type t struct{ n int }
		println(t{4}.n)
	}

	for i := 0; i < 2; i++ {
		type t struct{ n int }
		println("for", t{i}.n)
	}

	if true {
		type t struct{ n int }
		println("if", t{7}.n)
	}

	switch 1 {
	case 1:
		type t struct{ n int }
		println("case", t{8}.n)
	}

	{
		type a = int
		var q a = 5
		println("alias", q)
	}

	// Same-named types declared in sibling blocks are distinct types.
	var a, b interface{}
	{
		type t int
		a = t(1)
	}
	{
		type t int
		b = t(1)
	}
	println("distinct", a == b)
	var c, d interface{}
	{
		type t int
		c = t(1)
		d = t(1)
	}
	println("same", c == d)
}

// Output:
// 4
// for 0
// for 1
// if 7
// case 8
// alias 5
// distinct false
// same true
