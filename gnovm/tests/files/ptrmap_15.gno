// PKGPATH: gno.land/r/ptr_map
package ptr_map

type MyStruct struct {
	Key *int
}

var (
	m  = map[*int]string{}
	i1 = MyStruct{Key: new(int)}
)

func init() {
	*i1.Key = 1 // Set the value of the pointer
	m[i1.Key] = "first key"
}

func main(cur realm,) {
	println(m[i1.Key]) // Output: first key
}

// Output:
// first key
