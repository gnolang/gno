// PKGPATH: gno.land/r/ptr_map
package ptr_map

type MyStruct struct {
	Name string
	Age  int
}

var (
	m  = map[*MyStruct]string{}
	i1 = &MyStruct{Name: "alice", Age: 2}
)

func init() {
	m[i1] = "first key"
	println(m[i1])
}

func main(cur realm,) {
	i2 := *i1
	println(m[&i2] == "")

	i1.Age = 3
	println(m[i1])
}

// Output:
// first key
// true
// first key
