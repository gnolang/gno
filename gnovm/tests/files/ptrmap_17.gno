// PKGPATH: gno.land/r/ptr_map
package ptr_map

type Foo struct {
	name string
}

var (
	arr   = [3]Foo{Foo{"a"}, Foo{"b"}, Foo{"c"}}
	m     = map[*Foo]string{}
	index = 0
)

func init() {
	m[&arr[index]] = "first key"
}

func main(cur realm,) {
	println(m[&arr[index]])
}

// Output:
// first key
