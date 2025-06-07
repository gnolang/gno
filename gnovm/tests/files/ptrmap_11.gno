// PKGPATH: gno.land/r/ptr_map
package ptr_map

var (
	m     = map[*int]string{}
	a, b  = 1, 2
	S     = []*int{&a, &b} // slice
	index = 0
)

func init() {
	m[S[index]] = "first key"
}

func main(cur realm,) {
	println(m[S[index]]) // Output: first key
	println(m[S[1]] == "")
}

// Output:
// first key
// true
