// PKGPATH: gno.land/r/ptr_map
package ptr_map

type MyStruct struct {
	Name string
	Age  int
}

var (
	m = map[[1]*MyStruct]string{}

	i1 = &MyStruct{Name: "alice", Age: 2}
)

func main() {
	// Create an array of pointers to MyStruct
	key1 := [1]*MyStruct{i1}
	m[key1] = "first key"
	println(m[key1])
	m[key1] = "second key"
	println(m[key1])
}

// Output:
// first key
// second key
