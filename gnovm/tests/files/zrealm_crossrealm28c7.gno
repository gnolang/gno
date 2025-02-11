// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var s []*foo

func main() {
	var arr [3]*foo

	var f1 = foo{name: "1"}
	var f2 = foo{name: "2"}
	var f3 = foo{name: "3"}
	arr[0] = &f1
	arr[1] = &f2
	arr[2] = &f3

	s = arr[1:] // will attach the whole array
	println("ok")
}

// Output:
// ok
