// PKGPATH: gno.land/r/ptr_map
package ptr_map

type Foo struct {
	name string
}

type MyStruct struct {
	Name string
	Age  int
	key  Foo
}

var (
	m  = map[*Foo]string{}
	i1 = MyStruct{Name: "alice", Age: 2, key: Foo{name: "bob"}}
)

func init() {
	m[&i1.key] = "first key"
}

func main(cur realm,) {
	println(m[&i1.key])
}

// Output:
// first key
