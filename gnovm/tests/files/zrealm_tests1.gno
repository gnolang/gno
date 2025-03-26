// PKGPATH: gno.land/r/demo/tests
package tests

type Stringer interface {
	String() string
}

type Foo struct{ name string }

func (f *Foo) String() string {
	return "..."
}

var stringers []Stringer

func AddStringer(str Stringer) {
	stringers = append(stringers, str)
}

func init() {
	f1 := &Foo{name: "foo1"}
	AddStringer(f1)
}

func main() {
	f2 := &Foo{name: "foo2"}
	AddStringer(f2)
}
