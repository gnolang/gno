// PKGPATH: gno.land/r/main
package main

type A struct {
	b *B
}

type B struct {
	c *C
}

type C struct {
	d *D
}

type D struct {
	e int
}

var x = A{b: &B{c: &C{d: &D{e: 1}}}}

func init() {
	println("init() START")
	// variant of struct63.gno
	// x.b.c = C{d: D{e: 2}}
	x.b.c = &C{d: nil} // Is realm.go "created" C{}.
	x.b.c.d = &D{e: 2} // *Not* realm.go "updated" because C{} is unreal...
	// however it *will* get saved recursively due to "created".
	// If new-real or dirty any new objects in subcomponents are separately
	// marked "created" there, not here.
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
