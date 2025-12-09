// PKGPATH: gno.land/r/main
package main

type A struct {
	b B
}

type B struct {
	c C
}

type C struct {
	d D
}

type D struct {
	e int
}

var x = A{b: B{c: C{d: D{e: 1}}}}

func init() {
	println("init() START")
	x.b.c = C{d: D{e: 2}}
	println("init() END")
}

func main() {
	println("main() START")
	println(x.b.c.d.e)
	println("main() END")
}

// Output:
// init() START
// init() END
// main() START
// 2
// main() END
