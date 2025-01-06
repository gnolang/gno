// PKGPATH: gno.land/r/ptr_map
package ptr_map

type Foo struct {
	name string
}

var root *Foo

var f = &Foo{name: "a"}
var f2 = &Foo{name: "a2"}

func init() {
	root = f // origin of root is set as f's
}

func main() {
	f.name = "b"
	println(root == f)

	root = f2 // root has new origin(and pointer value)
	println(root == f2)
}

// Output:
// true
// true
