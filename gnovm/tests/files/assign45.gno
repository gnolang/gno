package main

// Names the func-decl refusal must not swallow: a package var holding a
// func value, and a local shadowing a package func.
func f() { println("f") }

var g = func() { println("g") }

func main() {
	g = f
	g()
	f := 1
	f = 2
	println(f)
}

// Output:
// f
// 2
