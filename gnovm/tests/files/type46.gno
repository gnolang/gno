package main

// Type declarations in every kind of block parsed from a file: else-if and
// else bodies, type-switch clauses, range bodies, labeled blocks, blocks in
// func literals, methods, defers, and a package-level var's initializer.
type R struct{}

func (R) M() int {
	{
		type t struct{ n int }
		return t{11}.n
	}
}

var pv = func() int {
	{
		type t struct{ n int }
		return t{12}.n
	}
}()

func main() {
	x := 2
	if x == 1 {
		type t struct{ n int }
		println("if", t{1}.n)
	} else if x == 2 {
		type t struct{ n int }
		println("else-if", t{2}.n)
	} else {
		type t struct{ n int }
		println("else", t{3}.n)
	}
	if x == 1 {
	} else {
		type t struct{ n int }
		println("else2", t{4}.n)
	}
	var v any = 1
	switch v.(type) {
	case int:
		type t struct{ n int }
		println("typeswitch", t{5}.n)
	default:
		type t struct{ s string }
		println(t{"no"}.s)
	}
	for k := range map[string]int{"a": 1} {
		type t struct{ s string }
		println("range", t{k}.s)
	}
L:
	{
		type t struct{ n int }
		println("labeled", t{6}.n)
		if false {
			goto L
		}
	}
	{
		f := func() int {
			{
				type t struct{ n int }
				return t{7}.n
			}
		}
		println("funclit", f())
	}
	{
		{
			{
				type t struct{ n int }
				println("deep", t{8}.n)
			}
		}
	}
	defer func() {
		{
			type t struct{ n int }
			println("defer", t{9}.n)
		}
	}()
	println("method", R{}.M())
	println("pkgvar", pv)
}

// Output:
// else-if 2
// else2 4
// typeswitch 5
// range a
// labeled 6
// funclit 7
// deep 8
// method 11
// pkgvar 12
// defer 9
