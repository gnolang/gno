// PKGPATH: gno.land/r/ptr_map
package ptr_map

var (
	m    = map[*int]string{}
	a, b = 1, 2
	S    = []*int{&a, &b}
)

func init() {
	m[S[0]] = "first key"
}

func main(cur realm,) {
	println(m[S[0]]) // Output: first key
	println(m[S[1]] == "")
}

// Output:
// first key
// true
