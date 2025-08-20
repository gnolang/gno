// PKGPATH: gno.land/r/ptr_map
package ptr_map

var (
	m         = map[*int]string{}
	a         = 0
	ptr *int  = &a // A pointer to an int
	i1  **int = &ptr
)

func init() {
	m[*i1] = "first key"
}

func main(cur realm,) {
	println(m[*i1]) // Output: first key
}

// Output:
// first key
