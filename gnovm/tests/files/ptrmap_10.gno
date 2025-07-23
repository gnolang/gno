// PKGPATH: gno.land/r/ptr_map
package ptr_map

var (
	m    = map[*int]string{}
	a, b = 1, 2
	arr  = [2]*int{&a, &b}
)

func init() {
	m[arr[0]] = "first key"
}

func main(cur realm,) {
	println(m[arr[0]]) // Output: first key
	println(m[arr[1]] == "")
}

// Output:
// first key
// true
