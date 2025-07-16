// PKGPATH: gno.land/r/ptr_map
package ptr_map

type Foo struct {
	name string
}

var (
	arr = [3]Foo{Foo{"a"}, Foo{"b"}, Foo{"c"}}
	m   = map[*Foo]string{}
)

func init() {
	m[&arr[0]] = "first key"
}

func main(cur realm,) {
	println(m[&arr[0]])
	println(m[&arr[1]] == "")
}

// Output:
// first key
// true
