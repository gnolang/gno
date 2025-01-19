// MAXALLOC: 25000
// max total allocation of 100 mb.

// PKGPATH: gno.land/r/test
package test

type Foo struct{ name string }

var f = Foo{name: "foo"}

func main() {
	println(f) // here should alloc amino for f
}
