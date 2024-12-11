// PKGPATH: gno.land/r/ptr_map
package ptr_map

type Foo struct {
	name string
}

var m = make(map[Foo]string)

func main() {

	i1 := Foo{
		name: "i1",
	}

	m[i1] = "a"

	println(m[i1])

	i2 := i1

	i1.name = "i2"

	println(m[i1] == "")

	println(m[i2])
}

// Output:
// a
// true
// a
