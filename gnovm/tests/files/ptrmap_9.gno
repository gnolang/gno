// PKGPATH: gno.land/r/ptr_map
package ptr_map

var m = make(map[*int]string)
var m2 = make(map[int]int)

var s []*int
var s2 []*int

func init() {
	s = append(s, new(int))
	m2[0] = 0
	// s[m2[0]] is pointer value,
	m[s[m2[0]]] = "a"
	s2 = append(s[:1], s[1:]...)
}

func main(cur realm,) {
	println(m[s[m2[0]]])
	println(s[m2[0]] == s2[m2[0]])
	println(m[s[m2[0]]] == m[s2[m2[0]]])
}

// Output:
// a
// true
// true
