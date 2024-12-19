// PKGPATH: gno.land/r/owner
package owner

type Foo struct{ name string }

type Bar struct {
	f    Foo
	name string
}

var root1 interface{}
var bar = Bar{name: "bar"}

func init() {
	bar.f = Foo{name: "f1"}
}

func main() {
	root1 = bar.f
	println(root1)
}

// Output:
// (struct{("f1" string)} gno.land/r/owner.Foo)
