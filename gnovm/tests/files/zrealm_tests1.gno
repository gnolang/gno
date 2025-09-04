// PKGPATH: gno.land/r/tests/vm
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

func main() {
	f1 := &Foo{name: "foo1"}
	AddStringer(f1)
	f2 := &Foo{name: "foo2"}
	AddStringer(f2)
	println("ok")
}

// Error:
// expected package name [vm] but got [tests]

// TypeCheckError:
// gno.land/r/tests/vm/.gnobuiltins.gno:1:1: package vm; expected package tests
