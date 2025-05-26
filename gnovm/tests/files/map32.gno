// PKGPATH: gno.land/r/map_test
package map_test

type Foo struct{ name string }

type Bar struct {
	name string
	f    Foo
	p    *int
}

var b Bar

func main() {
	f := Foo{name: "foo"}
	b.f = f

	i := 3
	i2 := &i
	b.p = i2 // attach...
}
