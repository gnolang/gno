// PKGPATH: gno.land/r/ptr_map
package ptr_map

type Foo struct {
	name string
}

var m = make(map[*Foo]string)

func main() {

	m[&Foo{name: "i1"}] = "a"

	println(m[&Foo{
		name: "i1",
	}] == "")
}

// Output:
// true
