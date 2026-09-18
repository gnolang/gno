package main

// A function type or interface method is a type, not a binding: its parameter
// and result names declare nothing, so the declaration check does not run on
// them. Only FuncDecl and FuncLitExpr names are checked.

type I interface{ M(len int) (cap int) }

var f func(len int) (cap int)

func main() {
	var i I
	_ = i
	_ = f
	println("ok")
}

// Output:
// ok
