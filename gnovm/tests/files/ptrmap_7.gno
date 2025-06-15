// PKGPATH: gno.land/r/ptr_map
package ptr_map

type S struct {
	i int
}

var m = make(map[*byte]string) // Initialize the map
var sArr = make([]*byte, 0, 4) // Use a slice of pointers
var a, b, c = byte('a'), byte('b'), byte('c')
var d, e, f = byte('d'), byte('e'), byte('f')

func init() {
	// Append pointers to the slice
	sArr = append(sArr, &a, &b, &c)
	m[sArr[1]] = "ok"
	println(&b == sArr[1])
}

func main(cur realm,) {
	// Create a new slice without reallocating memory for existing elements
	newArr := append(sArr[:1], sArr[2:]...)

	newArr = append(newArr, &d)
	newArr = append(newArr, &e)
	// a, c, d, e, f
	newArr = append(newArr, &f) // reallocation

	// Compare pointers directly
	println(sArr[1] == newArr[1])
	println(m[sArr[1]] == m[newArr[1]])
	println(m[newArr[1]] == "") // underlying base array changed
}

// Output:
// true
// true
// true
// true
